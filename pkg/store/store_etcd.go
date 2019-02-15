package store

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/fagongzi/gateway/pkg/client"
	pbutil "github.com/fagongzi/gateway/pkg/pb"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
	"github.com/fagongzi/gateway/pkg/route"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/util/format"
	"golang.org/x/net/context"
)

var (
	// ErrHasBind error has bind into, can not delete
	ErrHasBind = errors.New("Has bind info, can not delete")
	// ErrStaleOP is a stale error
	ErrStaleOP = errors.New("stale option")
)

const (
	// DefaultTimeout default timeout
	DefaultTimeout = time.Second * 3
	// DefaultRequestTimeout default request timeout
	DefaultRequestTimeout = 10 * time.Second
	// DefaultSlowRequestTime default slow request time
	DefaultSlowRequestTime = time.Second * 1

	batch = uint64(1000)
	endID = uint64(math.MaxUint64)
)

// EtcdStore etcd store impl
type EtcdStore struct {
	sync.RWMutex

	prefix      string
	clustersDir string
	serversDir  string
	bindsDir    string
	apisDir     string
	proxiesDir  string
	routingsDir string
	idPath      string

	idLock sync.Mutex
	base   uint64
	end    uint64

	evtCh              chan *Evt
	watchMethodMapping map[EvtSrc]func(EvtType, *mvccpb.KeyValue) *Evt

	rawClient *clientv3.Client
}

// NewEtcdStore create a etcd store
func NewEtcdStore(etcdAddrs []string, prefix string, basicAuth BasicAuth) (Store, error) {
	store := &EtcdStore{
		prefix:             prefix,
		clustersDir:        fmt.Sprintf("%s/clusters", prefix),
		serversDir:         fmt.Sprintf("%s/servers", prefix),
		bindsDir:           fmt.Sprintf("%s/binds", prefix),
		apisDir:            fmt.Sprintf("%s/apis", prefix),
		proxiesDir:         fmt.Sprintf("%s/proxies", prefix),
		routingsDir:        fmt.Sprintf("%s/routings", prefix),
		idPath:             fmt.Sprintf("%s/id", prefix),
		watchMethodMapping: make(map[EvtSrc]func(EvtType, *mvccpb.KeyValue) *Evt),
		base:               100,
		end:                100,
	}

	config := &clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: DefaultTimeout,
	}
	if basicAuth.userName != "" {
		config.Username = basicAuth.userName
	}
	if basicAuth.password != "" {
		config.Password = basicAuth.password
	}

	cli, err := clientv3.New(*config)

	if err != nil {
		return nil, err
	}

	store.rawClient = cli

	store.init()
	return store, nil
}

// Raw returns the raw client
func (e *EtcdStore) Raw() interface{} {
	return e.rawClient
}

// AddBind bind a server to a cluster
func (e *EtcdStore) AddBind(bind *metapb.Bind) error {
	e.Lock()
	defer e.Unlock()

	data, err := bind.Marshal()
	if err != nil {
		return err
	}

	return e.put(e.getBindKey(bind), string(data))
}

// Batch batch update
func (e *EtcdStore) Batch(batch *rpcpb.BatchReq) (*rpcpb.BatchRsp, error) {
	e.Lock()
	defer e.Unlock()

	rsp := &rpcpb.BatchRsp{}
	ops := make([]clientv3.Op, 0, len(batch.PutServers))
	for _, req := range batch.PutServers {
		value := &req.Server
		err := pbutil.ValidateServer(value)
		if err != nil {
			return nil, err
		}

		op, err := e.putPBWithOp(e.serversDir, value, func(id uint64) {
			value.ID = id
			rsp.PutServers = append(rsp.PutServers, &rpcpb.PutServerRsp{
				ID: id,
			})
		})
		if err != nil {
			return nil, err
		}

		ops = append(ops, op)
	}
	err := e.putBatch(ops...)
	if err != nil {
		return nil, err
	}

	ops = make([]clientv3.Op, 0, len(batch.PutClusters))
	for _, req := range batch.PutClusters {
		value := &req.Cluster
		err := pbutil.ValidateCluster(value)
		if err != nil {
			return nil, err
		}

		op, err := e.putPBWithOp(e.clustersDir, value, func(id uint64) {
			value.ID = id
			rsp.PutClusters = append(rsp.PutClusters, &rpcpb.PutClusterRsp{
				ID: id,
			})
		})
		if err != nil {
			return nil, err
		}

		ops = append(ops, op)
	}
	err = e.putBatch(ops...)
	if err != nil {
		return nil, err
	}

	ops = make([]clientv3.Op, 0, len(batch.AddBinds))
	for _, req := range batch.AddBinds {
		value := &metapb.Bind{
			ClusterID: req.Cluster,
			ServerID:  req.Server,
		}

		data, err := value.Marshal()
		if err != nil {
			return nil, err
		}

		ops = append(ops, e.op(e.getBindKey(value), string(data)))
		rsp.AddBinds = append(rsp.AddBinds, &rpcpb.AddBindRsp{})
	}

	err = e.putBatch(ops...)
	if err != nil {
		return nil, err
	}

	ops = make([]clientv3.Op, 0, len(batch.PutAPIs))
	for _, req := range batch.PutAPIs {
		value := &req.API
		err := pbutil.ValidateAPI(value)
		if err != nil {
			return nil, err
		}

		op, err := e.putPBWithOp(e.apisDir, value, func(id uint64) {
			value.ID = id
			rsp.PutAPIs = append(rsp.PutAPIs, &rpcpb.PutAPIRsp{
				ID: id,
			})
		})
		if err != nil {
			return nil, err
		}

		ops = append(ops, op)
	}
	err = e.putBatch(ops...)
	if err != nil {
		return nil, err
	}

	ops = make([]clientv3.Op, 0, len(batch.PutRoutings))
	for _, req := range batch.PutRoutings {
		value := &req.Routing
		err := pbutil.ValidateRouting(value)
		if err != nil {
			return nil, err
		}

		op, err := e.putPBWithOp(e.routingsDir, value, func(id uint64) {
			value.ID = id
			rsp.PutRoutings = append(rsp.PutRoutings, &rpcpb.PutRoutingRsp{
				ID: id,
			})
		})
		if err != nil {
			return nil, err
		}

		ops = append(ops, op)
	}
	err = e.putBatch(ops...)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// RemoveBind remove bind
func (e *EtcdStore) RemoveBind(bind *metapb.Bind) error {
	e.Lock()
	defer e.Unlock()

	return e.delete(e.getBindKey(bind))
}

// RemoveClusterBind remove cluster all bind servers
func (e *EtcdStore) RemoveClusterBind(id uint64) error {
	e.Lock()
	defer e.Unlock()

	return e.delete(e.getClusterBindPrefix(id), clientv3.WithPrefix())
}

// GetBindServers return cluster binds servers
func (e *EtcdStore) GetBindServers(id uint64) ([]uint64, error) {
	e.RLock()
	defer e.RUnlock()

	return e.doGetBindServers(id)
}

func (e *EtcdStore) doGetBindServers(id uint64) ([]uint64, error) {
	rsp, err := e.get(e.getClusterBindPrefix(id), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	if len(rsp.Kvs) == 0 {
		return nil, nil
	}

	var values []uint64
	for _, item := range rsp.Kvs {
		v := &metapb.Bind{}
		err := v.Unmarshal(item.Value)
		if err != nil {
			return nil, err
		}

		values = append(values, v.ServerID)
	}

	return values, nil
}

// PutCluster add or update the cluster
func (e *EtcdStore) PutCluster(value *metapb.Cluster) (uint64, error) {
	e.Lock()
	defer e.Unlock()

	err := pbutil.ValidateCluster(value)
	if err != nil {
		return 0, err
	}

	return e.putPB(e.clustersDir, value, func(id uint64) {
		value.ID = id
	})
}

// RemoveCluster remove the cluster and it's binds
func (e *EtcdStore) RemoveCluster(id uint64) error {
	e.Lock()
	defer e.Unlock()

	opCluster := clientv3.OpDelete(getKey(e.clustersDir, id))
	opBind := clientv3.OpDelete(e.getClusterBindPrefix(id), clientv3.WithPrefix())
	_, err := e.txn().Then(opCluster, opBind).Commit()
	return err
}

// GetClusters returns all clusters
func (e *EtcdStore) GetClusters(limit int64, fn func(interface{}) error) error {
	e.RLock()
	defer e.RUnlock()

	return e.getValues(e.clustersDir, limit, func() pb { return &metapb.Cluster{} }, fn)
}

// GetCluster returns the cluster
func (e *EtcdStore) GetCluster(id uint64) (*metapb.Cluster, error) {
	e.RLock()
	defer e.RUnlock()

	value := &metapb.Cluster{}
	return value, e.getPB(e.clustersDir, id, value)
}

// PutServer add or update the server
func (e *EtcdStore) PutServer(value *metapb.Server) (uint64, error) {
	e.Lock()
	defer e.Unlock()

	err := pbutil.ValidateServer(value)
	if err != nil {
		return 0, err
	}

	return e.putPB(e.serversDir, value, func(id uint64) {
		value.ID = id
	})
}

// RemoveServer remove the server
func (e *EtcdStore) RemoveServer(id uint64) error {
	e.Lock()
	defer e.Unlock()

	return e.delete(getKey(e.serversDir, id))
}

// GetServers returns all server
func (e *EtcdStore) GetServers(limit int64, fn func(interface{}) error) error {
	e.RLock()
	defer e.RUnlock()

	return e.getValues(e.serversDir, limit, func() pb { return &metapb.Server{} }, fn)
}

// GetServer returns the server
func (e *EtcdStore) GetServer(id uint64) (*metapb.Server, error) {
	e.RLock()
	defer e.RUnlock()

	value := &metapb.Server{}
	return value, e.getPB(e.serversDir, id, value)
}

// PutAPI add or update a API
func (e *EtcdStore) PutAPI(value *metapb.API) (uint64, error) {
	err := pbutil.ValidateAPI(value)
	if err != nil {
		return 0, err
	}

	e.Lock()
	defer e.Unlock()

	// load all api every times for validate
	// TODO: maybe need optimization if there are too much apis
	apiRoute := route.NewRoute()
	e.getValues(e.apisDir, 64, func() pb { return &metapb.API{} }, func(value interface{}) error {
		apiRoute.Add(value.(*metapb.API))
		return nil
	})

	err = apiRoute.Add(value)
	if err != nil {
		return 0, err
	}

	return e.putPB(e.apisDir, value, func(id uint64) {
		value.ID = id
	})
}

// RemoveAPI remove a api from store
func (e *EtcdStore) RemoveAPI(id uint64) error {
	e.Lock()
	defer e.Unlock()

	return e.delete(getKey(e.apisDir, id))
}

// GetAPIs returns all api
func (e *EtcdStore) GetAPIs(limit int64, fn func(interface{}) error) error {
	e.RLock()
	defer e.RUnlock()

	return e.getValues(e.apisDir, limit, func() pb { return &metapb.API{} }, fn)
}

// GetAPI returns the api
func (e *EtcdStore) GetAPI(id uint64) (*metapb.API, error) {
	e.RLock()
	defer e.RUnlock()

	value := &metapb.API{}
	return value, e.getPB(e.apisDir, id, value)
}

// PutRouting add or update routing
func (e *EtcdStore) PutRouting(value *metapb.Routing) (uint64, error) {
	e.Lock()
	defer e.Unlock()

	err := pbutil.ValidateRouting(value)
	if err != nil {
		return 0, err
	}

	return e.putPB(e.routingsDir, value, func(id uint64) {
		value.ID = id
	})
}

// RemoveRouting remove routing
func (e *EtcdStore) RemoveRouting(id uint64) error {
	e.Lock()
	defer e.Unlock()

	return e.delete(getKey(e.routingsDir, id))
}

// GetRoutings returns routes in store
func (e *EtcdStore) GetRoutings(limit int64, fn func(interface{}) error) error {
	e.RLock()
	defer e.RUnlock()

	return e.getValues(e.routingsDir, limit, func() pb { return &metapb.Routing{} }, fn)
}

// GetRouting returns a routing
func (e *EtcdStore) GetRouting(id uint64) (*metapb.Routing, error) {
	e.RLock()
	defer e.RUnlock()

	value := &metapb.Routing{}
	return value, e.getPB(e.routingsDir, id, value)
}

// RegistryProxy registry
func (e *EtcdStore) RegistryProxy(proxy *metapb.Proxy, ttl int64) error {
	key := getAddrKey(e.proxiesDir, proxy.Addr)
	data, err := proxy.Marshal()
	if err != nil {
		return err
	}

	lessor := clientv3.NewLease(e.rawClient)
	defer lessor.Close()

	ctx, cancel := context.WithTimeout(e.rawClient.Ctx(), DefaultRequestTimeout)
	leaseResp, err := lessor.Grant(ctx, ttl)
	cancel()
	if err != nil {
		return err
	}

	_, err = e.rawClient.KeepAlive(e.rawClient.Ctx(), leaseResp.ID)
	if err != nil {
		return err
	}

	return e.put(key, string(data), clientv3.WithLease(leaseResp.ID))
}

// GetProxies returns proxies in store
func (e *EtcdStore) GetProxies(limit int64, fn func(*metapb.Proxy) error) error {
	start := util.MinAddrFormat
	end := getAddrKey(e.proxiesDir, util.MaxAddrFormat)
	withRange := clientv3.WithRange(end)
	withLimit := clientv3.WithLimit(limit)

	for {
		resp, err := e.get(getAddrKey(e.proxiesDir, start), withRange, withLimit)
		if err != nil {
			return err
		}

		for _, item := range resp.Kvs {
			value := &metapb.Proxy{}
			err := value.Unmarshal(item.Value)
			if err != nil {
				return err
			}

			fn(value)

			start = util.GetAddrNextFormat(value.Addr)
		}

		// read complete
		if len(resp.Kvs) < int(limit) {
			break
		}
	}

	return nil
}

// Clean clean data in store
func (e *EtcdStore) Clean() error {
	e.Lock()
	defer e.Unlock()

	return e.delete(e.prefix, clientv3.WithPrefix())
}

// SetID set id
func (e *EtcdStore) SetID(id uint64) error {
	e.Lock()
	defer e.Unlock()

	op := clientv3.OpPut(e.idPath, string(format.Uint64ToBytes(id)))
	rsp, err := e.txn().Then(op).Commit()
	if err != nil {
		return err
	}

	if !rsp.Succeeded {
		return ErrStaleOP
	}

	e.end = 0
	e.base = 0
	return nil
}

// BackupTo backup to other gateway
func (e *EtcdStore) BackupTo(to string) error {
	e.Lock()
	defer e.Unlock()

	targetC, err := client.NewClient(time.Second*10, to)
	if err != nil {
		return err
	}

	defer targetC.Close()

	// Clean
	err = targetC.Clean()
	if err != nil {
		return err
	}

	limit := int64(96)
	batch := &rpcpb.BatchReq{}

	// backup server
	err = e.getValues(e.serversDir, limit, func() pb { return &metapb.Server{} }, func(value interface{}) error {
		batch.PutServers = append(batch.PutServers, &rpcpb.PutServerReq{
			Server: *value.(*metapb.Server),
		})

		if int64(len(batch.PutServers)) == limit {
			_, err := targetC.Batch(batch)
			if err != nil {
				return err
			}

			batch = &rpcpb.BatchReq{}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if int64(len(batch.PutServers)) > 0 {
		_, err := targetC.Batch(batch)
		if err != nil {
			return err
		}
	}

	// backup cluster
	batch = &rpcpb.BatchReq{}
	err = e.getValues(e.clustersDir, limit, func() pb { return &metapb.Cluster{} }, func(value interface{}) error {
		batch.PutClusters = append(batch.PutClusters, &rpcpb.PutClusterReq{
			Cluster: *value.(*metapb.Cluster),
		})

		if int64(len(batch.PutClusters)) == limit {
			_, err := targetC.Batch(batch)
			if err != nil {
				return err
			}

			batch = &rpcpb.BatchReq{}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if int64(len(batch.PutClusters)) > 0 {
		_, err := targetC.Batch(batch)
		if err != nil {
			return err
		}
	}

	// backup binds
	batch = &rpcpb.BatchReq{}
	err = e.getValues(e.clustersDir, limit, func() pb { return &metapb.Cluster{} }, func(value interface{}) error {
		cid := value.(*metapb.Cluster).ID
		servers, err := e.doGetBindServers(cid)
		if err != nil {
			return err
		}

		for _, sid := range servers {
			batch.AddBinds = append(batch.AddBinds, &rpcpb.AddBindReq{
				Cluster: cid,
				Server:  sid,
			})

			if int64(len(batch.AddBinds)) == limit {
				_, err := targetC.Batch(batch)
				if err != nil {
					return err
				}

				batch = &rpcpb.BatchReq{}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if int64(len(batch.AddBinds)) > 0 {
		_, err := targetC.Batch(batch)
		if err != nil {
			return err
		}

		batch = &rpcpb.BatchReq{}
	}

	// backup apis
	batch = &rpcpb.BatchReq{}
	err = e.getValues(e.apisDir, limit, func() pb { return &metapb.API{} }, func(value interface{}) error {
		batch.PutAPIs = append(batch.PutAPIs, &rpcpb.PutAPIReq{
			API: *value.(*metapb.API),
		})

		if int64(len(batch.PutAPIs)) == limit {
			_, err := targetC.Batch(batch)
			if err != nil {
				return err
			}

			batch = &rpcpb.BatchReq{}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if int64(len(batch.PutAPIs)) > 0 {
		_, err := targetC.Batch(batch)
		if err != nil {
			return err
		}
	}

	// backup routings
	batch = &rpcpb.BatchReq{}
	err = e.getValues(e.routingsDir, limit, func() pb { return &metapb.Routing{} }, func(value interface{}) error {
		batch.PutRoutings = append(batch.PutRoutings, &rpcpb.PutRoutingReq{
			Routing: *value.(*metapb.Routing),
		})

		if int64(len(batch.PutRoutings)) == limit {
			_, err := targetC.Batch(batch)
			if err != nil {
				return err
			}

			batch = &rpcpb.BatchReq{}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if int64(len(batch.PutRoutings)) > 0 {
		_, err := targetC.Batch(batch)
		if err != nil {
			return err
		}
	}

	// backup id
	currID, err := e.getID()
	if err != nil {
		return err
	}

	return targetC.SetID(currID)
}

// System returns system info
func (e *EtcdStore) System() (*metapb.System, error) {
	e.RLock()
	defer e.RUnlock()

	value := &metapb.System{}
	rsp, err := e.get(e.apisDir, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return nil, err
	}
	value.Count.API = rsp.Count

	rsp, err = e.get(e.clustersDir, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return nil, err
	}
	value.Count.Cluster = rsp.Count

	rsp, err = e.get(e.serversDir, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return nil, err
	}
	value.Count.Server = rsp.Count

	rsp, err = e.get(e.routingsDir, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return nil, err
	}
	value.Count.Routing = rsp.Count

	return value, nil
}

func (e *EtcdStore) put(key, value string, opts ...clientv3.OpOption) error {
	_, err := e.txn().Then(clientv3.OpPut(key, value, opts...)).Commit()
	return err
}

func (e *EtcdStore) op(key, value string, opts ...clientv3.OpOption) clientv3.Op {
	return clientv3.OpPut(key, value, opts...)
}

func (e *EtcdStore) putBatch(ops ...clientv3.Op) error {
	if len(ops) == 0 {
		return nil
	}

	_, err := e.txn().Then(ops...).Commit()
	return err
}

func (e *EtcdStore) putTTL(key, value string, ttl int64) error {
	lessor := clientv3.NewLease(e.rawClient)
	defer lessor.Close()

	ctx, cancel := context.WithTimeout(e.rawClient.Ctx(), DefaultRequestTimeout)
	leaseResp, err := lessor.Grant(ctx, ttl)
	cancel()

	if err != nil {
		return err
	}

	_, err = e.txn().Then(clientv3.OpPut(key, value, clientv3.WithLease(leaseResp.ID))).Commit()
	return err
}

func (e *EtcdStore) delete(key string, opts ...clientv3.OpOption) error {
	_, err := e.txn().Then(clientv3.OpDelete(key, opts...)).Commit()
	return err
}

type pb interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	GetID() uint64
}

func (e *EtcdStore) putPB(prefix string, value pb, do func(uint64)) (uint64, error) {
	if value.GetID() == 0 {
		id, err := e.allocID()
		if err != nil {
			return 0, err
		}

		do(id)
	}

	data, err := value.Marshal()
	if err != nil {
		return 0, err
	}

	return value.GetID(), e.put(getKey(prefix, value.GetID()), string(data))
}

func (e *EtcdStore) putPBWithOp(prefix string, value pb, do func(uint64)) (clientv3.Op, error) {
	if value.GetID() == 0 {
		id, err := e.allocID()
		if err != nil {
			return clientv3.Op{}, err
		}

		do(id)
	}

	data, err := value.Marshal()
	if err != nil {
		return clientv3.Op{}, err
	}

	return e.op(getKey(prefix, value.GetID()), string(data)), nil
}

func (e *EtcdStore) getValues(prefix string, limit int64, factory func() pb, fn func(interface{}) error) error {
	start := uint64(0)
	end := getKey(prefix, endID)
	withRange := clientv3.WithRange(end)
	withLimit := clientv3.WithLimit(limit)

	for {
		resp, err := e.get(getKey(prefix, start), withRange, withLimit)
		if err != nil {
			return err
		}

		for _, item := range resp.Kvs {
			value := factory()
			err := value.Unmarshal(item.Value)
			if err != nil {
				return err
			}

			fn(value)

			start = value.GetID() + 1
		}

		// read complete
		if len(resp.Kvs) < int(limit) {
			break
		}
	}

	return nil
}

func (e *EtcdStore) get(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(e.rawClient.Ctx(), DefaultRequestTimeout)
	defer cancel()

	return clientv3.NewKV(e.rawClient).Get(ctx, key, opts...)
}

func (e *EtcdStore) getPB(prefix string, id uint64, value pb) error {
	data, err := e.getValue(getKey(prefix, id))
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("<%d> not found", id)
	}

	err = value.Unmarshal(data)
	if err != nil {
		return err
	}

	return nil
}

func (e *EtcdStore) getValue(key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(e.rawClient.Ctx(), DefaultRequestTimeout)
	defer cancel()

	resp, err := clientv3.NewKV(e.rawClient).Get(ctx, key)
	if nil != err {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	return resp.Kvs[0].Value, nil
}

func (e *EtcdStore) allocID() (uint64, error) {
	e.idLock.Lock()
	defer e.idLock.Unlock()

	if e.base == e.end {
		end, err := e.generate()
		if err != nil {
			return 0, err
		}

		e.end = end
		e.base = e.end - batch
	}

	e.base++
	return e.base, nil
}

func (e *EtcdStore) generate() (uint64, error) {
	for {
		value, err := e.getID()
		if err != nil {
			return 0, err
		}

		max := value + batch

		// create id
		if value == 0 {
			max := value + batch
			err := e.createID(max)
			if err == ErrStaleOP {
				continue
			}
			if err != nil {
				return 0, err
			}

			return max, nil
		}

		err = e.updateID(value, max)
		if err == ErrStaleOP {
			continue
		}
		if err != nil {
			return 0, err
		}

		return max, nil
	}
}

func (e *EtcdStore) createID(value uint64) error {
	cmp := clientv3.Compare(clientv3.CreateRevision(e.idPath), "=", 0)
	op := clientv3.OpPut(e.idPath, string(format.Uint64ToBytes(value)))
	rsp, err := e.txn().If(cmp).Then(op).Commit()
	if err != nil {
		return err
	}

	if !rsp.Succeeded {
		return ErrStaleOP
	}

	return nil
}

func (e *EtcdStore) getID() (uint64, error) {
	value, err := e.getValue(e.idPath)
	if err != nil {
		return 0, err
	}

	if len(value) == 0 {
		return 0, nil
	}

	return format.BytesToUint64(value)
}

func (e *EtcdStore) updateID(old, value uint64) error {
	cmp := clientv3.Compare(clientv3.Value(e.idPath), "=", string(format.Uint64ToBytes(old)))
	op := clientv3.OpPut(e.idPath, string(format.Uint64ToBytes(value)))
	rsp, err := e.txn().If(cmp).Then(op).Commit()
	if err != nil {
		return err
	}

	if !rsp.Succeeded {
		return ErrStaleOP
	}

	return nil
}

func (e *EtcdStore) getClusterBindPrefix(id uint64) string {
	return getKey(e.bindsDir, id)
}

func (e *EtcdStore) getBindKey(bind *metapb.Bind) string {
	return getKey(e.getClusterBindPrefix(bind.ClusterID), bind.ServerID)
}

package store

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/util/format"
	"github.com/fagongzi/util/task"
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
	runner    *task.Runner
}

// NewEtcdStore create a etcd store
func NewEtcdStore(etcdAddrs []string, prefix string, taskRunner *task.Runner) (Store, error) {
	store := &EtcdStore{
		prefix:             prefix,
		clustersDir:        fmt.Sprintf("%s/clusters", prefix),
		serversDir:         fmt.Sprintf("%s/servers", prefix),
		bindsDir:           fmt.Sprintf("%s/binds", prefix),
		apisDir:            fmt.Sprintf("%s/apis", prefix),
		proxiesDir:         fmt.Sprintf("%s/proxy", prefix),
		routingsDir:        fmt.Sprintf("%s/routings", prefix),
		idPath:             fmt.Sprintf("%s/id", prefix),
		watchMethodMapping: make(map[EvtSrc]func(EvtType, *mvccpb.KeyValue) *Evt),
		runner:             taskRunner,
		base:               100,
		end:                100,
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: DefaultTimeout,
	})

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
	data, err := bind.Marshal()
	if err != nil {
		return err
	}

	return e.put(e.getBindKey(bind), string(data))
}

// RemoveBind remove bind
func (e *EtcdStore) RemoveBind(bind *metapb.Bind) error {
	return e.delete(e.getBindKey(bind))
}

// RemoveClusterBind remove cluster all bind servers
func (e *EtcdStore) RemoveClusterBind(id uint64) error {
	return e.delete(e.getClusterBindPrefix(id), clientv3.WithPrefix())
}

// GetBindServers return cluster binds servers
func (e *EtcdStore) GetBindServers(id uint64) ([]uint64, error) {
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
	return e.putPB(e.clustersDir, value, func(id uint64) {
		value.ID = id
	})
}

// RemoveCluster remove the cluster and it's binds
func (e *EtcdStore) RemoveCluster(id uint64) error {
	opCluster := clientv3.OpDelete(getKey(e.clustersDir, id))
	opBind := clientv3.OpDelete(e.getClusterBindPrefix(id), clientv3.WithPrefix())
	_, err := e.txn().Then(opCluster, opBind).Commit()
	return err
}

// GetClusters returns all clusters
func (e *EtcdStore) GetClusters(limit int64, fn func(interface{}) error) error {
	return e.getValues(e.clustersDir, limit, &metapb.Cluster{}, fn)
}

// GetCluster returns the cluster
func (e *EtcdStore) GetCluster(id uint64) (*metapb.Cluster, error) {
	value := &metapb.Cluster{}
	return value, e.getPB(e.clustersDir, id, value)
}

// PutServer add or update the server
func (e *EtcdStore) PutServer(value *metapb.Server) (uint64, error) {
	return e.putPB(e.serversDir, value, func(id uint64) {
		value.ID = id
	})
}

// RemoveServer remove the server
func (e *EtcdStore) RemoveServer(id uint64) error {
	return e.delete(getKey(e.serversDir, id))
}

// GetServers returns all server
func (e *EtcdStore) GetServers(limit int64, fn func(interface{}) error) error {
	return e.getValues(e.serversDir, limit, &metapb.Server{}, fn)
}

// GetServer returns the server
func (e *EtcdStore) GetServer(id uint64) (*metapb.Server, error) {
	value := &metapb.Server{}
	return value, e.getPB(e.serversDir, id, value)
}

// PutAPI add or update a API
func (e *EtcdStore) PutAPI(value *metapb.API) (uint64, error) {
	return e.putPB(e.apisDir, value, func(id uint64) {
		value.ID = id
	})
}

// RemoveAPI remove a api from store
func (e *EtcdStore) RemoveAPI(id uint64) error {
	return e.delete(getKey(e.apisDir, id))
}

// GetAPIs returns all api
func (e *EtcdStore) GetAPIs(limit int64, fn func(interface{}) error) error {
	return e.getValues(e.apisDir, limit, &metapb.API{}, fn)
}

// GetAPI returns the api
func (e *EtcdStore) GetAPI(id uint64) (*metapb.API, error) {
	value := &metapb.API{}
	return value, e.getPB(e.apisDir, id, value)
}

// PutRouting add or update routing
func (e *EtcdStore) PutRouting(value *metapb.Routing) (uint64, error) {
	return e.putPB(e.routingsDir, value, func(id uint64) {
		value.ID = id
	})
}

// RemoveRouting remove routing
func (e *EtcdStore) RemoveRouting(id uint64) error {
	return e.delete(getKey(e.routingsDir, id))
}

// GetRoutings returns routes in store
func (e *EtcdStore) GetRoutings(limit int64, fn func(interface{}) error) error {
	return e.getValues(e.routingsDir, limit, &metapb.Routing{}, fn)
}

// GetRouting returns a routing
func (e *EtcdStore) GetRouting(id uint64) (*metapb.Routing, error) {
	value := &metapb.Routing{}
	return value, e.getPB(e.routingsDir, id, value)
}

// Clean clean data in store
func (e *EtcdStore) Clean() error {
	return e.delete(e.prefix, clientv3.WithPrefix())
}

func (e *EtcdStore) put(key, value string, opts ...clientv3.OpOption) error {
	_, err := e.txn().Then(clientv3.OpPut(key, value, opts...)).Commit()
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

func (e *EtcdStore) getValues(prefix string, limit int64, value pb, fn func(interface{}) error) error {
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

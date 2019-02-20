package service

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
	"github.com/fagongzi/gateway/pkg/store"
	"golang.org/x/net/context"
)

const (
	limit = int64(32)
)

type metaService struct {
	db store.Store
}

func newMetaService(db store.Store) rpcpb.MetaServiceServer {
	return &metaService{
		db: db,
	}
}

func (s *metaService) PutCluster(ctx context.Context, req *rpcpb.PutClusterReq) (*rpcpb.PutClusterRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		id, err := s.db.PutCluster(&req.Cluster)
		if err != nil {
			return nil, err
		}

		return &rpcpb.PutClusterRsp{
			ID: id,
		}, nil
	}
}

func (s *metaService) RemoveCluster(ctx context.Context, req *rpcpb.RemoveClusterReq) (*rpcpb.RemoveClusterRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		err := s.db.RemoveCluster(req.ID)
		if err != nil {
			return nil, err
		}

		return &rpcpb.RemoveClusterRsp{}, nil
	}
}

func (s *metaService) GetCluster(ctx context.Context, req *rpcpb.GetClusterReq) (*rpcpb.GetClusterRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		value, err := s.db.GetCluster(req.ID)
		if err != nil {
			return nil, err
		}

		return &rpcpb.GetClusterRsp{
			Cluster: value,
		}, nil
	}
}

func (s *metaService) GetClusterList(req *rpcpb.GetClusterListReq, stream rpcpb.MetaService_GetClusterListServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return errRPCCancel
		default:
			err := s.db.GetClusters(limit, func(value interface{}) error {
				return stream.Send(value.(*metapb.Cluster))
			})

			if err != nil {
				return err
			}

			return nil
		}
	}
}

func (s *metaService) PutServer(ctx context.Context, req *rpcpb.PutServerReq) (*rpcpb.PutServerRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		id, err := s.db.PutServer(&req.Server)
		if err != nil {
			return nil, err
		}

		return &rpcpb.PutServerRsp{
			ID: id,
		}, nil
	}
}

func (s *metaService) RemoveServer(ctx context.Context, req *rpcpb.RemoveServerReq) (*rpcpb.RemoveServerRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		err := s.db.RemoveServer(req.ID)
		if err != nil {
			return nil, err
		}

		return &rpcpb.RemoveServerRsp{}, nil
	}
}

func (s *metaService) GetServer(ctx context.Context, req *rpcpb.GetServerReq) (*rpcpb.GetServerRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		value, err := s.db.GetServer(req.ID)
		if err != nil {
			return nil, err
		}

		return &rpcpb.GetServerRsp{
			Server: value,
		}, nil
	}
}

func (s *metaService) GetServerList(req *rpcpb.GetServerListReq, stream rpcpb.MetaService_GetServerListServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return errRPCCancel
		default:
			err := s.db.GetServers(limit, func(value interface{}) error {
				return stream.Send(value.(*metapb.Server))
			})

			if err != nil {
				return err
			}

			return nil
		}
	}
}

func (s *metaService) PutAPI(ctx context.Context, req *rpcpb.PutAPIReq) (*rpcpb.PutAPIRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		id, err := s.db.PutAPI(&req.API)
		if err != nil {
			return nil, err
		}

		return &rpcpb.PutAPIRsp{
			ID: id,
		}, nil
	}
}

func (s *metaService) RemoveAPI(ctx context.Context, req *rpcpb.RemoveAPIReq) (*rpcpb.RemoveAPIRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		err := s.db.RemoveAPI(req.ID)
		if err != nil {
			return nil, err
		}

		return &rpcpb.RemoveAPIRsp{}, nil
	}
}

func (s *metaService) GetAPI(ctx context.Context, req *rpcpb.GetAPIReq) (*rpcpb.GetAPIRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		value, err := s.db.GetAPI(req.ID)
		if err != nil {
			return nil, err
		}

		return &rpcpb.GetAPIRsp{
			API: value,
		}, nil
	}
}

func (s *metaService) GetAPIList(req *rpcpb.GetAPIListReq, stream rpcpb.MetaService_GetAPIListServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return errRPCCancel
		default:
			err := s.db.GetAPIs(limit, func(value interface{}) error {
				return stream.Send(value.(*metapb.API))
			})

			if err != nil {
				return err
			}

			return nil
		}
	}
}

func (s *metaService) PutRouting(ctx context.Context, req *rpcpb.PutRoutingReq) (*rpcpb.PutRoutingRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		id, err := s.db.PutRouting(&req.Routing)
		if err != nil {
			return nil, err
		}

		return &rpcpb.PutRoutingRsp{
			ID: id,
		}, nil
	}
}

func (s *metaService) RemoveRouting(ctx context.Context, req *rpcpb.RemoveRoutingReq) (*rpcpb.RemoveRoutingRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		err := s.db.RemoveRouting(req.ID)
		if err != nil {
			return nil, err
		}

		return &rpcpb.RemoveRoutingRsp{}, nil
	}
}

func (s *metaService) GetRouting(ctx context.Context, req *rpcpb.GetRoutingReq) (*rpcpb.GetRoutingRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		value, err := s.db.GetRouting(req.ID)
		if err != nil {
			return nil, err
		}

		return &rpcpb.GetRoutingRsp{
			Routing: value,
		}, nil
	}
}

func (s *metaService) GetRoutingList(req *rpcpb.GetRoutingListReq, stream rpcpb.MetaService_GetRoutingListServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return errRPCCancel
		default:
			err := s.db.GetRoutings(limit, func(value interface{}) error {
				return stream.Send(value.(*metapb.Routing))
			})

			if err != nil {
				return err
			}

			return nil
		}
	}
}

func (s *metaService) AddBind(ctx context.Context, req *rpcpb.AddBindReq) (*rpcpb.AddBindRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		err := s.db.AddBind(&metapb.Bind{
			ClusterID: req.Cluster,
			ServerID:  req.Server,
		})
		if err != nil {
			return nil, err
		}

		return &rpcpb.AddBindRsp{}, nil
	}
}

func (s *metaService) Batch(ctx context.Context, req *rpcpb.BatchReq) (*rpcpb.BatchRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		return s.db.Batch(req)
	}
}

func (s *metaService) RemoveBind(ctx context.Context, req *rpcpb.RemoveBindReq) (*rpcpb.RemoveBindRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		err := s.db.RemoveBind(&metapb.Bind{
			ClusterID: req.Cluster,
			ServerID:  req.Server,
		})
		if err != nil {
			return nil, err
		}

		return &rpcpb.RemoveBindRsp{}, nil
	}
}

func (s *metaService) RemoveClusterBind(ctx context.Context, req *rpcpb.RemoveClusterBindReq) (*rpcpb.RemoveClusterBindRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		err := s.db.RemoveClusterBind(req.Cluster)
		if err != nil {
			return nil, err
		}

		return &rpcpb.RemoveClusterBindRsp{}, nil
	}
}

func (s *metaService) GetBindServers(ctx context.Context, req *rpcpb.GetBindServersReq) (*rpcpb.GetBindServersRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		servers, err := s.db.GetBindServers(req.Cluster)
		if err != nil {
			return nil, err
		}

		return &rpcpb.GetBindServersRsp{
			Servers: servers,
		}, nil
	}
}

func (s *metaService) PutPlugin(ctx context.Context, req *rpcpb.PutPluginReq) (*rpcpb.PutPluginRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		id, err := s.db.PutPlugin(&req.Plugin)
		if err != nil {
			return nil, err
		}

		return &rpcpb.PutPluginRsp{
			ID: id,
		}, nil
	}
}

func (s *metaService) RemovePlugin(ctx context.Context, req *rpcpb.RemovePluginReq) (*rpcpb.RemovePluginRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		err := s.db.RemovePlugin(req.ID)
		if err != nil {
			return nil, err
		}

		return &rpcpb.RemovePluginRsp{}, nil
	}
}

func (s *metaService) GetPlugin(ctx context.Context, req *rpcpb.GetPluginReq) (*rpcpb.GetPluginRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		value, err := s.db.GetPlugin(req.ID)
		if err != nil {
			return nil, err
		}

		return &rpcpb.GetPluginRsp{
			Plugin: value,
		}, nil
	}
}

func (s *metaService) GetPluginList(req *rpcpb.GetPluginListReq, stream rpcpb.MetaService_GetPluginListServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return errRPCCancel
		default:
			err := s.db.GetPlugins(limit, func(value interface{}) error {
				return stream.Send(value.(*metapb.Plugin))
			})

			if err != nil {
				return err
			}

			return nil
		}
	}
}

func (s *metaService) ApplyPlugins(ctx context.Context, req *rpcpb.ApplyPluginsReq) (*rpcpb.ApplyPluginsRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		err := s.db.ApplyPlugins(req.Applied)
		if err != nil {
			return nil, err
		}

		return &rpcpb.ApplyPluginsRsp{}, nil
	}
}

func (s *metaService) GetAppliedPlugins(ctx context.Context, req *rpcpb.GetAppliedPluginsReq) (*rpcpb.GetAppliedPluginsRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		value, err := s.db.GetAppliedPlugins()
		if err != nil {
			return nil, err
		}

		return &rpcpb.GetAppliedPluginsRsp{
			Applied: value,
		}, nil
	}
}

func (s *metaService) Clean(ctx context.Context, req *rpcpb.CleanReq) (*rpcpb.CleanRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		err := s.db.Clean()
		if err != nil {
			return nil, err
		}

		return &rpcpb.CleanRsp{}, nil
	}
}

func (s *metaService) SetID(ctx context.Context, req *rpcpb.SetIDReq) (*rpcpb.SetIDRsp, error) {
	select {
	case <-ctx.Done():
		return nil, errRPCCancel
	default:
		err := s.db.SetID(req.ID)
		if err != nil {
			return nil, err
		}

		return &rpcpb.SetIDRsp{}, nil
	}
}

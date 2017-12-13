package apiserver

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

func newMetaService(db store.Store) *metaService {
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

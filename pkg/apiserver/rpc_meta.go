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

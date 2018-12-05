package service

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/grpcx"
	"github.com/fagongzi/log"
	"github.com/labstack/echo"
)

func initClusterRouter(server *echo.Group) {
	server.GET("/clusters/:id",
		grpcx.NewGetHTTPHandle(idParamFactory, getClusterHandler))
	server.GET("/clusters/:id/binds",
		grpcx.NewGetHTTPHandle(idParamFactory, bindsClusterHandler))
	server.DELETE("/clusters/:id",
		grpcx.NewGetHTTPHandle(idParamFactory, deleteClusterHandler))
	server.DELETE("/clusters/:id/binds",
		grpcx.NewGetHTTPHandle(idParamFactory, deleteClusterBindsHandler))
	server.PUT("/clusters",
		grpcx.NewJSONBodyHTTPHandle(putClusterFactory, postClusterHandler))
	server.GET("/clusters",
		grpcx.NewGetHTTPHandle(limitQueryFactory, listClusterHandler))
}

func postClusterHandler(value interface{}) (*grpcx.JSONResult, error) {
	id, err := Store.PutCluster(value.(*metapb.Cluster))
	if err != nil {
		log.Errorf("api-cluster-put: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: id}, nil
}

func deleteClusterHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.RemoveCluster(value.(uint64))
	if err != nil {
		log.Errorf("api-cluster-delete: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{}, nil
}

func deleteClusterBindsHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.RemoveClusterBind(value.(uint64))
	if err != nil {
		log.Errorf("api-cluster-binds-delete: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{}, nil
}

func getClusterHandler(value interface{}) (*grpcx.JSONResult, error) {
	value, err := Store.GetCluster(value.(uint64))
	if err != nil {
		log.Errorf("api-cluster-get: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: value}, nil
}

func bindsClusterHandler(value interface{}) (*grpcx.JSONResult, error) {
	values, err := Store.GetBindServers(value.(uint64))
	if err != nil {
		log.Errorf("api-cluster-binds-get: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: values}, nil
}

func listClusterHandler(value interface{}) (*grpcx.JSONResult, error) {
	query := value.(*limitQuery)
	var values []*metapb.Cluster

	err := Store.GetClusters(limit, func(data interface{}) error {
		v := data.(*metapb.Cluster)
		if int64(len(values)) < query.limit && v.ID > query.afterID {
			values = append(values, v)
		}
		return nil
	})
	if err != nil {
		log.Errorf("api-cluster-list-get: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: values}, nil
}

func putClusterFactory() interface{} {
	return &metapb.Cluster{}
}

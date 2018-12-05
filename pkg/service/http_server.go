package service

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/grpcx"
	"github.com/fagongzi/log"
	"github.com/labstack/echo"
)

func initServerRouter(server *echo.Group) {
	server.GET("/servers/:id",
		grpcx.NewGetHTTPHandle(idParamFactory, getServerHandler))
	server.DELETE("/servers/:id",
		grpcx.NewGetHTTPHandle(idParamFactory, deleteServerHandler))
	server.PUT("/servers",
		grpcx.NewJSONBodyHTTPHandle(putServerFactory, postServerHandler))
	server.GET("/servers",
		grpcx.NewGetHTTPHandle(limitQueryFactory, listServerHandler))
}

func postServerHandler(value interface{}) (*grpcx.JSONResult, error) {
	id, err := Store.PutServer(value.(*metapb.Server))
	if err != nil {
		log.Errorf("api-server-put: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: id}, nil
}

func deleteServerHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.RemoveServer(value.(uint64))
	if err != nil {
		log.Errorf("api-server-delete: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{}, nil
}

func getServerHandler(value interface{}) (*grpcx.JSONResult, error) {
	value, err := Store.GetServer(value.(uint64))
	if err != nil {
		log.Errorf("api-server-get: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: value}, nil
}

func listServerHandler(value interface{}) (*grpcx.JSONResult, error) {
	query := value.(*limitQuery)
	var values []*metapb.Server

	err := Store.GetServers(limit, func(data interface{}) error {
		v := data.(*metapb.Server)
		if int64(len(values)) < query.limit && v.ID > query.afterID {
			values = append(values, v)
		}
		return nil
	})
	if err != nil {
		log.Errorf("api-server-list-get: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: values}, nil
}

func putServerFactory() interface{} {
	return &metapb.Server{}
}

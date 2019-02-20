package service

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/grpcx"
	"github.com/fagongzi/log"
	"github.com/labstack/echo"
)

func initPluginRouter(server *echo.Group) {
	server.GET("/plugins/:id",
		grpcx.NewGetHTTPHandle(idParamFactory, getPluginHandler))
	server.DELETE("/plugins/:id",
		grpcx.NewGetHTTPHandle(idParamFactory, deletePluginHandler))
	server.PUT("/plugins",
		grpcx.NewJSONBodyHTTPHandle(putPluginFactory, postPluginHandler))
	server.GET("/plugins",
		grpcx.NewGetHTTPHandle(limitQueryFactory, listPluginHandler))
}

func postPluginHandler(value interface{}) (*grpcx.JSONResult, error) {
	id, err := Store.PutPlugin(value.(*metapb.Plugin))
	if err != nil {
		log.Errorf("api-plugin-put: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: id}, nil
}

func deletePluginHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.RemovePlugin(value.(uint64))
	if err != nil {
		log.Errorf("api-plugin-delete: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{}, nil
}

func getPluginHandler(value interface{}) (*grpcx.JSONResult, error) {
	value, err := Store.GetPlugin(value.(uint64))
	if err != nil {
		log.Errorf("api-plugin-get: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: value}, nil
}

func listPluginHandler(value interface{}) (*grpcx.JSONResult, error) {
	query := value.(*limitQuery)
	var values []*metapb.Plugin

	err := Store.GetPlugins(limit, func(data interface{}) error {
		v := data.(*metapb.Plugin)
		if int64(len(values)) < query.limit && v.ID > query.afterID {
			values = append(values, v)
		}
		return nil
	})
	if err != nil {
		log.Errorf("api-plugin-list-get: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: values}, nil
}

func putPluginFactory() interface{} {
	return &metapb.Plugin{}
}

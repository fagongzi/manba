package service

import (
	"fmt"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/grpcx"
	"github.com/fagongzi/log"
	"github.com/labstack/echo"
)

func initRoutingRouter(server *echo.Echo) {
	server.GET(fmt.Sprintf("/%s%s", apiVersion, "/routings/:id"),
		grpcx.NewGetHTTPHandle(idParamFactory, getRoutingHandler))
	server.DELETE(fmt.Sprintf("/%s%s", apiVersion, "/routings/:id"),
		grpcx.NewGetHTTPHandle(idParamFactory, deleteRoutingHandler))
	server.PUT(fmt.Sprintf("/%s%s", apiVersion, "/routings"),
		grpcx.NewJSONBodyHTTPHandle(putRoutingFactory, postRoutingHandler))
	server.GET(fmt.Sprintf("/%s%s", apiVersion, "/routings"),
		grpcx.NewGetHTTPHandle(limitQueryFactory, listRoutingHandler))
}

func postRoutingHandler(value interface{}) (*grpcx.JSONResult, error) {
	id, err := Store.PutRouting(value.(*metapb.Routing))
	if err != nil {
		log.Errorf("api-routing-put: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: id}, nil
}

func deleteRoutingHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.RemoveRouting(value.(uint64))
	if err != nil {
		log.Errorf("api-routing-delete: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{}, nil
}

func getRoutingHandler(value interface{}) (*grpcx.JSONResult, error) {
	value, err := Store.GetRouting(value.(uint64))
	if err != nil {
		log.Errorf("api-routing-delete: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: value}, nil
}

func putRoutingFactory() interface{} {
	return &metapb.Routing{}
}

func listRoutingHandler(value interface{}) (*grpcx.JSONResult, error) {
	query := value.(*limitQuery)
	var values []*metapb.Routing

	err := Store.GetRoutings(limit, func(data interface{}) error {
		v := data.(*metapb.Routing)
		if int64(len(values)) < query.limit && v.ID > query.afterID {
			values = append(values, v)
		}
		return nil
	})
	if err != nil {
		log.Errorf("api-routing-list-get: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: values}, nil
}

package service

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/grpcx"
	"github.com/fagongzi/log"
	"github.com/labstack/echo"
)

func initRoutingRouter(server *echo.Group) {
	server.GET("/routings/:id",
		grpcx.NewGetHTTPHandle(idParamFactory, getRoutingHandler))
	server.DELETE("/routings/:id",
		grpcx.NewGetHTTPHandle(idParamFactory, deleteRoutingHandler))
	server.PUT("/routings",
		grpcx.NewJSONBodyHTTPHandle(putRoutingFactory, postRoutingHandler))
	server.GET("/routings",
		grpcx.NewGetHTTPHandle(limitQueryFactory, listRoutingHandler))
}

func postRoutingHandler(value interface{}) (*grpcx.JSONResult, error) {
	id, err := Store.PutRouting(value.(*metapb.Routing))
	if err != nil {
		log.Errorf("api-routing-put: req %+v, errors:%+v", value, err)
		return &grpcx.JSONResult{Code: -1, Data: err.Error()}, nil
	}

	return &grpcx.JSONResult{Data: id}, nil
}

func deleteRoutingHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.RemoveRouting(value.(uint64))
	if err != nil {
		log.Errorf("api-routing-delete: req %+v, errors:%+v", value, err)
		return &grpcx.JSONResult{Code: -1, Data: err.Error()}, nil
	}

	return &grpcx.JSONResult{}, nil
}

func getRoutingHandler(value interface{}) (*grpcx.JSONResult, error) {
	value, err := Store.GetRouting(value.(uint64))
	if err != nil {
		log.Errorf("api-routing-delete: req %+v, errors:%+v", value, err)
		return &grpcx.JSONResult{Code: -1, Data: err.Error()}, nil
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
		return &grpcx.JSONResult{Code: -1, Data: err.Error()}, nil
	}

	return &grpcx.JSONResult{Data: values}, nil
}

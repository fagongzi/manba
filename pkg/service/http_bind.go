package service

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/grpcx"
	"github.com/fagongzi/log"
	"github.com/labstack/echo"
)

func initBindRouter(server *echo.Group) {
	server.DELETE("/binds",
		grpcx.NewJSONBodyHTTPHandle(bindFactory, deleteBindHandler))

	server.PUT("/binds",
		grpcx.NewJSONBodyHTTPHandle(bindFactory, postBindHandler))
}

func postBindHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.AddBind(value.(*metapb.Bind))
	if err != nil {
		log.Errorf("api-bind-put: req %+v, errors:%+v", value, err)
		return &grpcx.JSONResult{Code: -1, Data: err.Error()}, nil
	}

	return &grpcx.JSONResult{}, nil
}

func deleteBindHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.RemoveBind(value.(*metapb.Bind))
	if err != nil {
		log.Errorf("api-bind-delete: req %+v, errors:%+v", value, err)
		return &grpcx.JSONResult{Code: -1, Data: err.Error()}, nil
	}

	return &grpcx.JSONResult{}, nil
}

func bindFactory() interface{} {
	return &metapb.Bind{}
}

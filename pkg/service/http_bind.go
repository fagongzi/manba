package service

import (
	"fmt"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/grpcx"
	"github.com/fagongzi/log"
	"github.com/labstack/echo"
)

func initBindRouter(server *echo.Echo) {
	server.DELETE(fmt.Sprintf("/%s%s", apiVersion, "/binds"),
		grpcx.NewJSONBodyHTTPHandle(bindFactory, deleteBindHandler))

	server.PUT(fmt.Sprintf("/%s%s", apiVersion, "/binds"),
		grpcx.NewJSONBodyHTTPHandle(bindFactory, postBindHandler))
}

func postBindHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.AddBind(value.(*metapb.Bind))
	if err != nil {
		log.Errorf("api-bind-put: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{}, nil
}

func deleteBindHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.RemoveBind(value.(*metapb.Bind))
	if err != nil {
		log.Errorf("api-bind-delete: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{}, nil
}

func bindFactory() interface{} {
	return &metapb.Bind{}
}

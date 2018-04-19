package service

import (
	"fmt"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/grpcx"
	"github.com/fagongzi/log"
	"github.com/labstack/echo"
)

func initAPIRouter(server *echo.Echo) {
	server.GET(fmt.Sprintf("/%s%s", apiVersion, "/apis/:id"),
		grpcx.NewGetHTTPHandle(idParamFactory, getAPIHandler))
	server.DELETE(fmt.Sprintf("/%s%s", apiVersion, "/apis/:id"),
		grpcx.NewGetHTTPHandle(idParamFactory, deleteAPIHandler))
	server.PUT(fmt.Sprintf("/%s%s", apiVersion, "/apis"),
		grpcx.NewJSONBodyHTTPHandle(putAPIFactory, postAPIHandler))
	server.GET(fmt.Sprintf("/%s%s", apiVersion, "/apis"),
		grpcx.NewGetHTTPHandle(limitQueryFactory, listAPIHandler))
}

func postAPIHandler(value interface{}) (*grpcx.JSONResult, error) {
	id, err := Store.PutAPI(value.(*metapb.API))
	if err != nil {
		log.Errorf("api-api-put: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: id}, nil
}

func deleteAPIHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.RemoveAPI(value.(uint64))
	if err != nil {
		log.Errorf("api-api-delete: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{}, nil
}

func getAPIHandler(value interface{}) (*grpcx.JSONResult, error) {
	value, err := Store.GetAPI(value.(uint64))
	if err != nil {
		log.Errorf("api-api-get: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: value}, nil
}

func listAPIHandler(value interface{}) (*grpcx.JSONResult, error) {
	query := value.(*limitQuery)
	var values []*metapb.API

	err := Store.GetAPIs(limit, func(data interface{}) error {
		v := data.(*metapb.API)
		if int64(len(values)) < query.limit && v.ID > query.afterID {
			values = append(values, v)
		}
		return nil
	})
	if err != nil {
		log.Errorf("api-api-list-get: req %+v, errors:%+v", value, err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: values}, nil
}

func putAPIFactory() interface{} {
	return &metapb.API{}
}

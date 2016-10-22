package server

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
)

func (server *AdminServer) getAggregations() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		aggregations, err := server.store.GetAggregations()
		if err != nil {
			errstr = err.Error()
			code = CodeError
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: aggregations,
		})
	}
}

func (server *AdminServer) getAggregation() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		fmt.Println(c.QueryParam("url"))
		u, _ := url.QueryUnescape(c.QueryParam("url"))
		fmt.Println(u)
		aggregation, err := server.store.GetAggregation(u)
		if err != nil {
			errstr = err.Error()
			code = CodeError
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: aggregation,
		})
	}
}

func (server *AdminServer) newAggregation() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		ang, err := model.UnMarshalAggregationFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := server.store.SaveAggregation(ang)
			if nil != err {
				errstr = err.Error()
				code = CodeError
			}
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
		})
	}
}

func (server *AdminServer) updateAggregation() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		ang, err := model.UnMarshalAggregationFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := server.store.UpdateAggregation(ang)
			if nil != err {
				errstr = err.Error()
				code = CodeError
			}
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
		})
	}
}

func (server *AdminServer) deleteAggregation() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		url := c.QueryParam("url")
		err := server.store.DeleteAggregation(url)

		if nil != err {
			errstr = err.Error()
			code = CodeError
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
		})
	}
}

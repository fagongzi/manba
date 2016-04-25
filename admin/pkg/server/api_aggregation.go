package server

import (
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
	"net/http"
)

func (self *AdminServer) getAggregations() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		aggregations, err := self.store.GetAggregations()
		if err != nil {
			errstr = err.Error()
			code = CODE_ERROR
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: aggregations,
		})
	}
}

func (self *AdminServer) newAggregation() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		ang, err := model.UnMarshalAggregationFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		} else {
			err := self.store.SaveAggregation(ang)
			if nil != err {
				errstr = err.Error()
				code = CODE_ERROR
			}
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
		})
	}
}

func (self *AdminServer) deleteAggregation() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		url := c.QueryParam("url")
		err := self.store.DeleteAggregation(url)

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
		})
	}
}

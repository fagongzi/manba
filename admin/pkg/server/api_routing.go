package server

import (
	"fmt"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
	"net/http"
)

func (self *AdminServer) getRoutings() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		routings, err := self.store.GetRoutings()
		if err != nil {
			errstr = err.Error()
			code = CODE_ERROR
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: routings,
		})
	}
}

func (self *AdminServer) newRouting() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		routing, err := model.UnMarshalRoutingFromReader(c.Request().Body())

		fmt.Printf("%+v\n", routing)

		if err == nil {
			err = routing.Check()
		}

		fmt.Printf("%+v, %s\n", routing, err)

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		} else {
			err := self.store.SaveRouting(routing)
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

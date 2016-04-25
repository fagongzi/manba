package server

import (
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
	"net/http"
)

func (self *AdminServer) getProxies() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		registor, _ := self.store.(model.Register)

		proxies, err := registor.GetProxies()
		if err != nil {
			errstr = err.Error()
			code = CODE_ERROR
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: proxies,
		})
	}
}

func (self *AdminServer) changeLogLevel() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		addr := c.Param("addr")
		level := c.Param("level")

		registor, _ := self.store.(model.Register)

		err := registor.ChangeLogLevel(addr, level)

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

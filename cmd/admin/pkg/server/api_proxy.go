package server

import (
	"net/http"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
)

func (server *AdminServer) getProxies() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		registor, _ := server.store.(model.Register)

		proxies, err := registor.GetProxies()
		if err != nil {
			errstr = err.Error()
			code = CodeError
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: proxies,
		})
	}
}

func (server *AdminServer) changeLogLevel() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		addr := c.Param("addr")
		level := c.Param("level")

		registor, _ := server.store.(model.Register)

		err := registor.ChangeLogLevel(addr, level)

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

package server

import (
	"net/http"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
)

func (server *AdminServer) getServers() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		servers, err := server.store.GetServers()
		if err != nil {
			errstr = err.Error()
			code = CodeError
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: servers,
		})
	}
}

func (server *AdminServer) getServer() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		id := c.Param("id")
		server, err := server.store.GetServer(id, true)

		if nil != err {
			errstr = err.Error()
			code = CodeError
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: server,
		})
	}
}

func (server *AdminServer) updateServer() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		svr, err := model.UnMarshalServerFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := server.store.UpdateServer(svr)
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

func (server *AdminServer) newServer() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		svr, err := model.UnMarshalServerFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := server.store.SaveServer(svr)
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

func (server *AdminServer) deleteServer() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		id := c.Param("id")
		err := server.store.DeleteServer(id)

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

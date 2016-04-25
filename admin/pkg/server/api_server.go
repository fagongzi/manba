package server

import (
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
	"net/http"
)

func (self *AdminServer) getServers() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		servers, err := self.store.GetServers()
		if err != nil {
			errstr = err.Error()
			code = CODE_ERROR
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: servers,
		})
	}
}

func (self *AdminServer) getServer() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		id := c.Param("id")
		server, err := self.store.GetServer(id, true)

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: server,
		})
	}
}

func (self *AdminServer) updateServer() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		server, err := model.UnMarshalServerFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		} else {
			err := self.store.UpdateServer(server)
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

func (self *AdminServer) newServer() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		server, err := model.UnMarshalServerFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		} else {
			err := self.store.SaveServer(server)
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

func (self *AdminServer) deleteServer() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		id := c.Param("id")
		err := self.store.DeleteServer(id)

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

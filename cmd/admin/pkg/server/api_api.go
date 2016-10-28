package server

import (
	"net/http"

	"encoding/base64"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
)

func (server *AdminServer) getAPIs() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		apis, err := server.store.GetAPIs()
		if err != nil {
			errstr = err.Error()
			code = CodeError
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: apis,
		})
	}
}

func (server *AdminServer) getAPI() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		u, _ := base64.RawURLEncoding.DecodeString(c.Param("url"))
		method := c.QueryParam("method")

		api, err := server.store.GetAPI(string(u), method)
		if err != nil {
			errstr = err.Error()
			code = CodeError
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: api,
		})
	}
}

func (server *AdminServer) newAPI() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		api, err := model.UnMarshalAPIFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := server.store.SaveAPI(api)
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

func (server *AdminServer) updateAPI() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		ang, err := model.UnMarshalAPIFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := server.store.UpdateAPI(ang)
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

func (server *AdminServer) deleteAPI() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		url, _ := base64.RawURLEncoding.DecodeString(c.Param("url"))
		method := c.QueryParam("method")
		err := server.store.DeleteAPI(string(url), method)

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

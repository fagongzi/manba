package server

import (
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
	"net/http"
)

func (server *AdminServer) newBind() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		bind, err := model.UnMarshalBindFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := server.store.SaveBind(bind)
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

func (server *AdminServer) unBind() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		bind, err := model.UnMarshalBindFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := server.store.UnBind(bind)
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

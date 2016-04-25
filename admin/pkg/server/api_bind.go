package server

import (
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
	"net/http"
)

func (self *AdminServer) newBind() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		bind, err := model.UnMarshalBindFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		} else {
			err := self.store.SaveBind(bind)
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

func (self *AdminServer) unBind() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		bind, err := model.UnMarshalBindFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		} else {
			err := self.store.UnBind(bind)
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

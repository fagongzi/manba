package api

import (
	"encoding/base64"
	"net/http"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
)

func (s *Server) initAPIOfAPIs() {
	s.api.POST("/api/v1/apis", s.createAPI())
	s.api.DELETE("/api/v1/apis/:url", s.deleteAPI())
	s.api.PUT("/api/v1/apis/:url", s.updateAPI())
	s.api.GET("/api/v1/apis", s.listAPIs())
	s.api.GET("/api/v1/apis/:url", s.getAPI())
}

func (s *Server) listAPIs() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		apis, err := s.store.GetAPIs()
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

func (s *Server) getAPI() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		u, _ := base64.RawURLEncoding.DecodeString(c.Param("url"))
		method := c.QueryParam("method")

		api, err := s.store.GetAPI(string(u), method)
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

func (s *Server) createAPI() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		api, err := model.UnMarshalAPIFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := api.Validate()
			if err != nil {
				errstr = err.Error()
				code = CodeError
			} else {
				err = s.store.SaveAPI(api)
				if nil != err {
					errstr = err.Error()
					code = CodeError
				}
			}

		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
		})
	}
}

func (s *Server) updateAPI() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		api, err := model.UnMarshalAPIFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := api.Validate()
			if err != nil {
				errstr = err.Error()
				code = CodeError
			} else {
				err := s.store.UpdateAPI(api)
				if nil != err {
					errstr = err.Error()
					code = CodeError
				}
			}
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
		})
	}
}

func (s *Server) deleteAPI() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		url, _ := base64.RawURLEncoding.DecodeString(c.Param("url"))
		method := c.QueryParam("method")
		err := s.store.DeleteAPI(string(url), method)

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

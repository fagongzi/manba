package api

import (
	"net/http"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
)

func (s *Server) initAPIOfAPIs() {
	s.api.POST("/api/v1/apis", s.createAPI())
	s.api.DELETE("/api/v1/apis/:id", s.deleteAPI())
	s.api.PUT("/api/v1/apis/:id", s.updateAPI())
	s.api.GET("/api/v1/apis", s.listAPIs())
	s.api.GET("/api/v1/apis/:id", s.getAPI())
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

		api, err := s.store.GetAPI(c.Param("id"))
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

		api := &model.API{}
		err := readJSONFromReader(api, c.Request().Body)

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

		api := &model.API{}
		err := readJSONFromReader(api, c.Request().Body)

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

		err := s.store.DeleteAPI(c.Param("id"))

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

package api

import (
	"net/http"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
)

func (s *Server) initAPIOfRoutings() {
	s.api.POST("/api/v1/routings", s.createRouting())
	s.api.GET("/api/v1/routings", s.listRoutings())
}

func (s *Server) listRoutings() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		routings, err := s.store.GetRoutings()
		if err != nil {
			errstr = err.Error()
			code = CodeError
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: routings,
		})
	}
}

func (s *Server) createRouting() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		routing := &model.Routing{}
		err := readJSONFromReader(routing, c.Request().Body)
		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := routing.Validate()
			if err != nil {
				errstr = err.Error()
				code = CodeError
			} else {
				err := s.store.SaveRouting(routing)
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

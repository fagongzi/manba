package api

import (
	"net/http"

	"github.com/fagongzi/gateway/pkg/lb"
	"github.com/labstack/echo"
)

func (s *Server) initAPIOfLBS() {
	s.api.GET("/api/v1/lbs", s.listLBS())
}

func (s *Server) listLBS() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, &Result{
			Code:  CodeSuccess,
			Value: lb.GetSupportLBS(),
		})
	}
}

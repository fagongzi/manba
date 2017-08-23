package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
)

// AnalysisPoint analysis point
type AnalysisPoint struct {
	ProxyAddr  string `json:"proxyAddr,omitempty"`
	ServerAddr string `json:"serverAddr,omitempty"`
	Secs       int    `json:"secs,omitempty"`
}

func unmarshalAnalysisPointFromReader(r io.Reader) (*AnalysisPoint, error) {
	v := &AnalysisPoint{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	return v, err
}

func (s *Server) initAPIOfAnalysis() {
	s.api.GET("/api/v1/analysis/:proxy/:server/:secs", s.listAnalysis())
	s.api.POST("/api/v1/analysis", s.createAnalysis())
}

func (s *Server) listAnalysis() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		proxyAddr := c.Param("proxy")
		serverAddr := c.Param("server")
		secs, err := strconv.Atoi(c.Param("secs"))

		if nil != err {
			return c.JSON(http.StatusOK, &Result{
				Code:  code,
				Error: errstr,
			})
		}

		registor, _ := s.store.(model.Register)

		data, err := registor.GetAnalysisPoint(proxyAddr, serverAddr, secs)
		if err != nil {
			errstr = err.Error()
			code = CodeError
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: data,
		})
	}
}

func (s *Server) createAnalysis() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		point, err := unmarshalAnalysisPointFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			registor, _ := s.store.(model.Register)

			err := registor.AddAnalysisPoint(point.ProxyAddr, point.ServerAddr, point.Secs)
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

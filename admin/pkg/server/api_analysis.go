package server

import (
	"encoding/json"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
	"io"
	"net/http"
	"strconv"
)

type AnalysisPoint struct {
	ProxyAddr  string `json:"proxyAddr,omitempty"`
	ServerAddr string `json:"serverAddr,omitempty"`
	Secs       int    `json:"secs,omitempty"`
}

func unMarshalAnalysisPointFromReader(r io.Reader) (*AnalysisPoint, error) {
	v := &AnalysisPoint{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	return v, err
}

func (self *AdminServer) getAnalysis() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		proxyAddr := c.Param("proxy")
		serverAddr := c.Param("server")
		secs, err := strconv.Atoi(c.Param("secs"))

		if nil != err {
			return c.JSON(http.StatusOK, &Result{
				Code:  code,
				Error: errstr,
			})
		}

		registor, _ := self.store.(model.Register)

		data, err := registor.GetAnalysisPoint(proxyAddr, serverAddr, secs)
		if err != nil {
			errstr = err.Error()
			code = CODE_ERROR
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: data,
		})
	}
}

func (self *AdminServer) newAnalysis() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		point, err := unMarshalAnalysisPointFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		} else {
			registor, _ := self.store.(model.Register)

			err := registor.AddAnalysisPoint(point.ProxyAddr, point.ServerAddr, point.Secs)
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

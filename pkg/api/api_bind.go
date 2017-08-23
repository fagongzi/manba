package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
)

// Binds create binds
type Binds struct {
	Target  string   `json:"target"`
	Servers []string `json:"servers"`
}

func unMarshalBindsFromReader(r io.Reader) (*Binds, error) {
	v := &Binds{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	if nil != err {
		return nil, err
	}

	return v, nil
}

func (s *Server) initAPIOfBinds() {
	s.api.POST("/api/v1/binds", s.createBind())
	s.api.DELETE("/api/v1/binds", s.deleteBind())
}

func (s *Server) createBind() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		binds, err := unMarshalBindsFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			for _, addr := range binds.Servers {
				bind := &model.Bind{
					ClusterName: binds.Target,
					ServerAddr:  addr,
				}

				err := bind.Validate()
				if err != nil {
					errstr = err.Error()
					code = CodeError
					break
				} else {
					err := s.store.SaveBind(bind)
					if nil != err {
						errstr = err.Error()
						code = CodeError
						break
					}
				}
			}
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
		})
	}
}

func (s *Server) deleteBind() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		bind, err := model.UnMarshalBindFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := s.store.UnBind(bind)
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

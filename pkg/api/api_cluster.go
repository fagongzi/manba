package api

import (
	"net/http"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
)

func (s *Server) initAPIOfClusters() {
	s.api.POST("/api/clusters", s.createCluster())
	s.api.DELETE("/api/clusters/:id", s.deleteCluster())
	s.api.PUT("/api/clusters/:id", s.updateCluster())
	s.api.GET("/api/clusters/:id", s.getCluster())
	s.api.GET("/api/clusters", s.listClusters())
}

func (s *Server) listClusters() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		clusters, err := s.store.GetClusters()
		if err != nil {
			errstr = err.Error()
			code = CodeError
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: clusters,
		})
	}
}

func (s *Server) getCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		id := c.Param("id")
		cluster, err := s.store.GetCluster(id)

		if nil != err {
			errstr = err.Error()
			code = CodeError
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: cluster,
		})
	}
}

func (s *Server) createCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		cluster, err := model.UnMarshalClusterFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := s.store.SaveCluster(cluster)
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

func (s *Server) updateCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		cluster, err := model.UnMarshalClusterFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := s.store.UpdateCluster(cluster)
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

func (s *Server) deleteCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		id := c.Param("id")
		err := s.store.DeleteCluster(id)

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

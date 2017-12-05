package api

import (
	"net/http"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
)

func (s *Server) initAPIOfClusters() {
	s.api.POST("/api/v1/clusters", s.createCluster())
	s.api.DELETE("/api/v1/clusters/:id", s.deleteCluster())
	s.api.PUT("/api/v1/clusters/:id", s.updateCluster())
	s.api.GET("/api/v1/clusters/:id", s.getCluster())
	s.api.GET("/api/v1/clusters", s.listClusters())
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

		cluster := &model.Cluster{}
		err := readJSONFromReader(cluster, c.Request().Body)

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := cluster.Validate()
			if err != nil {
				errstr = err.Error()
				code = CodeError
			} else {
				err := s.store.SaveCluster(cluster)
				if nil != err {
					errstr = err.Error()
					code = CodeError
				}
			}
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: cluster.ID,
		})
	}
}

func (s *Server) updateCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		cluster := &model.Cluster{}
		err := readJSONFromReader(cluster, c.Request().Body)

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := cluster.Validate()
			if err != nil {
				errstr = err.Error()
				code = CodeError
			} else {
				err := s.store.UpdateCluster(cluster)
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

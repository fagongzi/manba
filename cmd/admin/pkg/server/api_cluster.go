package server

import (
	"net/http"

	"github.com/fagongzi/gateway/pkg/lb"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
)

func (server *AdminServer) getLbs() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, lb.GetSupportLBS())
	}
}

func (server *AdminServer) getCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		id := c.Param("id")
		cluster, err := server.store.GetCluster(id)

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

func (server *AdminServer) getClusters() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		clusters, err := server.store.GetClusters()
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

func (server *AdminServer) newCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		cluster, err := model.UnMarshalClusterFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := server.store.SaveCluster(cluster)
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

func (server *AdminServer) updateCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		cluster, err := model.UnMarshalClusterFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CodeError
		} else {
			err := server.store.UpdateCluster(cluster)
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

func (server *AdminServer) deleteCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CodeSuccess

		id := c.Param("id")
		err := server.store.DeleteCluster(id)

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

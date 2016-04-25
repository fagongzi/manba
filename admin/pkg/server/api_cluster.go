package server

import (
	"github.com/fagongzi/gateway/pkg/lb"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
	"net/http"
)

func (self *AdminServer) getLbs() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, lb.GetSupportLBS())
	}
}

func (self *AdminServer) getCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		id := c.Param("id")
		cluster, err := self.store.GetCluster(id, true)

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: cluster,
		})
	}
}

func (self *AdminServer) getClusters() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		clusters, err := self.store.GetClusters()
		if err != nil {
			errstr = err.Error()
			code = CODE_ERROR
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
			Value: clusters,
		})
	}
}

func (self *AdminServer) newCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		cluster, err := model.UnMarshalClusterFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		} else {
			err := self.store.SaveCluster(cluster)
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

func (self *AdminServer) updateCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		cluster, err := model.UnMarshalClusterFromReader(c.Request().Body())

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		} else {
			err := self.store.UpdateCluster(cluster)
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

func (self *AdminServer) deleteCluster() echo.HandlerFunc {
	return func(c echo.Context) error {
		var errstr string
		code := CODE_SUCCESS

		id := c.Param("id")
		err := self.store.DeleteCluster(id)

		if nil != err {
			errstr = err.Error()
			code = CODE_ERROR
		}

		return c.JSON(http.StatusOK, &Result{
			Code:  code,
			Error: errstr,
		})
	}
}

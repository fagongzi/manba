package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/fagongzi/util/format"
	"github.com/labstack/echo"
	md "github.com/labstack/echo/middleware"
)

var (
	addr = flag.String("addr", "127.0.0.1:9090", "addr for backend")
)

func main() {
	flag.Parse()

	server := echo.New()
	server.Use(md.Logger())

	server.GET("/fail", func(c echo.Context) error {
		sleep := c.QueryParam("sleep")
		if sleep != "" {
			time.Sleep(time.Second * time.Duration(format.MustParseStrInt(sleep)))
		}

		code := c.QueryParam("code")
		if code != "" {
			return c.String(format.MustParseStrInt(code), "OK")
		}

		return c.String(http.StatusOK, "OK")
	})

	server.GET("/check", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	server.GET("/header", func(c echo.Context) error {
		name := c.QueryParam("name")
		return c.String(http.StatusOK, c.Request().Header.Get(name))
	})

	server.GET("/host", func(c echo.Context) error {
		return c.String(http.StatusOK, c.Request().Host)
	})

	server.GET("/error", func(c echo.Context) error {
		return c.NoContent(http.StatusBadRequest)
	})

	server.GET("/v1/components/:id", func(c echo.Context) error {
		value := make(map[string]interface{})
		data := make(map[string]interface{})
		user := make(map[string]interface{})
		user["id"] = c.Param("id")
		user["name"] = fmt.Sprintf("v1-name-%s", c.Param("id"))
		data["user"] = user
		data["source"] = *addr

		value["code"] = "0"
		value["data"] = data
		return c.JSON(http.StatusOK, value)
	})
	server.GET("/v1/users/:id", func(c echo.Context) error {
		user := make(map[string]interface{})
		user["id"] = c.Param("id")
		user["name"] = fmt.Sprintf("v1-name-%s", c.Param("id"))
		user["source"] = *addr
		return c.JSON(http.StatusOK, user)
	})
	server.GET("/v1/account/:id", func(c echo.Context) error {
		account := make(map[string]interface{})
		account["id"] = c.Param("id")
		account["source"] = *addr
		account["account"] = fmt.Sprintf("v1-account-%s", c.Param("id"))
		return c.JSON(http.StatusOK, account)
	})

	server.GET("/v2/users/:id", func(c echo.Context) error {
		user := make(map[string]interface{})
		user["id"] = c.Param("id")
		user["source"] = *addr
		user["name"] = fmt.Sprintf("v2-name-%s", c.Param("id"))
		return c.JSON(http.StatusOK, user)
	})
	server.GET("/v2/account/:id", func(c echo.Context) error {
		account := make(map[string]interface{})
		account["id"] = c.Param("id")
		account["source"] = *addr
		account["account"] = fmt.Sprintf("v2-account-%s", c.Param("id"))
		return c.JSON(http.StatusOK, account)
	})

	server.Start(*addr)
}

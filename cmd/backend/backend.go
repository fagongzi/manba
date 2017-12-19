package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

var (
	addr = flag.String("addr", "127.0.0.1:9090", "addr for backend")
)

func main() {
	server := echo.New()
	server.GET("/check", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	server.GET("/v1/users/:id", func(c echo.Context) error {
		user := make(map[string]interface{})
		user["id"] = c.Param("id")
		user["name"] = fmt.Sprintf("v1-name-%s", c.Param("id"))
		return c.JSON(http.StatusOK, user)
	})
	server.GET("/v1/account/:id", func(c echo.Context) error {
		account := make(map[string]interface{})
		account["id"] = c.Param("id")
		account["account"] = fmt.Sprintf("v1-account-%s", c.Param("id"))
		return c.JSON(http.StatusOK, account)
	})

	server.GET("/v2/users/:id", func(c echo.Context) error {
		user := make(map[string]interface{})
		user["id"] = c.Param("id")
		user["name"] = fmt.Sprintf("v2-name-%s", c.Param("id"))
		return c.JSON(http.StatusOK, user)
	})
	server.GET("/v2/account/:id", func(c echo.Context) error {
		account := make(map[string]interface{})
		account["id"] = c.Param("id")
		account["account"] = fmt.Sprintf("v2-account-%s", c.Param("id"))
		return c.JSON(http.StatusOK, account)
	})

	server.Start(*addr)
}

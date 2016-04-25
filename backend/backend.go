package main

import (
	"flag"

	"github.com/labstack/echo"
	sd "github.com/labstack/echo/engine/standard"
	mw "github.com/labstack/echo/middleware"
	"net/http"
	"strconv"
)

var (
	addr = flag.String("addr", ":80", "listen addr.(e.g. ip:port)")
)

func main() {
	flag.Parse()

	e := echo.New()

	e.SetDebug(true)

	e.Use(mw.Logger())
	e.Use(mw.Recover())

	e.Get("/check", check())
	e.Get("/api/age", age())
	e.Get("/api/name", name())

	e.Run(sd.New(*addr))
}

func check() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}
}

func age() echo.HandlerFunc {
	return func(c echo.Context) error {
		t := c.QueryParam("code")

		if "" != t {
			code, _ := strconv.Atoi(t)
			return c.NoContent(code)
		}

		cookie := &http.Cookie{
			Name:  "age-cookie",
			Value: "18",
		}

		c.Response().Header().Add("Set-Cookie", cookie.String())
		c.Response().Header().Add("age", "100")
		c.Response().Header().Add("common", "2")
		return c.JSON(http.StatusOK, 100)
	}
}

func name() echo.HandlerFunc {
	return func(c echo.Context) error {
		t := c.QueryParam("code")

		if "" != t {
			code, _ := strconv.Atoi(t)
			return c.NoContent(code)
		}

		cookie := &http.Cookie{
			Name:  "name-cookie",
			Value: "zhangxu",
		}

		c.Response().Header().Add("Set-Cookie", cookie.String())
		c.Response().Header().Add("name", "zhangsan")
		c.Response().Header().Add("common", "3")
		return c.JSON(http.StatusOK, "zhangsan")
	}
}

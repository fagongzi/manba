package main

import (
	"flag"

	"github.com/labstack/echo"
	sd "github.com/labstack/echo/engine/standard"
	mw "github.com/labstack/echo/middleware"
	"net/http"
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
	e.Get("/api/call", call())

	e.Run(sd.New(*addr))
}

func check() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}
}

func call() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, *addr)
	}
}

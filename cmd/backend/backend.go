package main

import (
	"flag"
	"time"

	"net/http"

	"strconv"

	"github.com/labstack/echo"
	sd "github.com/labstack/echo/engine/standard"
	mw "github.com/labstack/echo/middleware"
)

var (
	addr = flag.String("addr", ":80", "listen addr.(e.g. ip:port)")
)

type Cookie struct {
	name     string
	value    string
	path     string
	domain   string
	expires  time.Time
	secure   bool
	httpOnly bool
}

func (c Cookie) Name() string {
	return c.name
}

func (c Cookie) Value() string {
	return c.value
}

func (c Cookie) Path() string {
	return c.path
}

func (c Cookie) Domain() string {
	return c.domain
}

func (c Cookie) Expires() time.Time {
	return c.expires
}

func (c Cookie) Secure() bool {
	return c.secure
}

func (c Cookie) HTTPOnly() bool {
	return c.httpOnly
}

func main() {
	flag.Parse()

	e := echo.New()

	e.SetDebug(true)

	e.Use(mw.Logger())
	e.Use(mw.Recover())

	e.Get("/check", check())
	e.Get("/api/call", call())
	e.Get("/api/wait", wait())
	e.Get("/api/cookie", cookie())
	e.Get("/api/set-cookie", setCookie())
	e.Get("/api/query", query())
	e.Get("/api/path/:value", path())

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

func wait() echo.HandlerFunc {
	return func(c echo.Context) error {
		sec := c.QueryParam("value")
		s, _ := strconv.Atoi(sec)
		time.Sleep(time.Second * time.Duration(s))
		return c.JSON(http.StatusOK, *addr)
	}
}

func cookie() echo.HandlerFunc {
	return func(c echo.Context) error {
		ck, _ := c.Cookie("value")
		c.SetCookie(ck)
		return c.JSON(http.StatusOK, ck.Value())
	}
}

func setCookie() echo.HandlerFunc {
	return func(c echo.Context) error {
		value := c.QueryParam("value")
		c.SetCookie(&Cookie{
			name:  "value",
			value: value,
			path:  "/",
		})
		return c.JSON(http.StatusOK, value)
	}
}

func query() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, c.QueryParam("value"))
	}
}

func path() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, c.Param("value"))
	}
}

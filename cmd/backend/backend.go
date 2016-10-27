package main

import (
	"flag"

	"net/http"

	"github.com/labstack/echo"
	sd "github.com/labstack/echo/engine/standard"
	mw "github.com/labstack/echo/middleware"
)

var (
	addr = flag.String("addr", ":80", "listen addr.(e.g. ip:port)")
)

// UserBase user base
type UserBase struct {
	UserID   int    `json:"userID"`
	UserName string `json:"userName"`
}

// UserAccount user account
type UserAccount struct {
	UserID int `json:"userID"`
	Money  int `json:"money"`
}

var (
	user = &UserBase{
		UserID:   1,
		UserName: "Owen",
	}

	account = &UserAccount{
		UserID: 1,
		Money:  100,
	}
)

func main() {
	flag.Parse()
	e := echo.New()
	e.Use(mw.Recover())

	e.GET("/check", check())
	e.GET("/api/users/:userId/base", userBase())
	e.GET("/api/users/:userId/account", userAccount())

	e.Run(sd.New(*addr))
}

func check() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}
}

func userBase() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, user)
	}
}

func userAccount() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, userAccount)
	}
}

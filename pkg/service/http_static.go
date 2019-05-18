package service

import (
	"github.com/labstack/echo"
)

func initStatic(server *echo.Echo, ui, uiPrefix string) {
	server.Static(uiPrefix, ui)
}

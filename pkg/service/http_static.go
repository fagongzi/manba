package service

import (
	"github.com/labstack/echo"
)

func initStatic(server *echo.Group, ui, uiPrefix string) {
	server.Static(uiPrefix, ui)
}

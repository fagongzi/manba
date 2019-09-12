package service

import (
	"fmt"

	"github.com/labstack/echo"
)

func initStatic(server *echo.Echo, ui, uiPrefix string) {
	server.Static(uiPrefix, ui)
	server.Static("static", fmt.Sprintf("%s/static", ui))
}

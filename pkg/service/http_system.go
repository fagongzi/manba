package service

import (
	"fmt"

	"github.com/fagongzi/grpcx"
	"github.com/fagongzi/log"
	"github.com/labstack/echo"
)

type backup struct {
	ToAddr string `json:"toAddr"`
}

func initSystemRouter(server *echo.Echo) {
	server.GET(fmt.Sprintf("/%s%s", apiVersion, "/system"),
		grpcx.NewGetHTTPHandle(emptyParamFactory, getSystemHandler))

	server.POST(fmt.Sprintf("/%s%s", apiVersion, "/system/backup"),
		grpcx.NewJSONBodyHTTPHandle(backupFactory, postBackupHandler))
}

func getSystemHandler(value interface{}) (*grpcx.JSONResult, error) {
	info, err := Store.System()
	if err != nil {
		log.Errorf("api-system-get: errors:%+v", err)
		return nil, err
	}

	return &grpcx.JSONResult{Data: info}, nil
}

func postBackupHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.BackupTo(value.(*backup).ToAddr)
	if err != nil {
		log.Errorf("api-system-backup: errors:%+v", err)
		return nil, err
	}

	return &grpcx.JSONResult{}, nil
}

func backupFactory() interface{} {
	return &backup{}
}

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

func initBackupRouter(server *echo.Echo) {
	server.POST(fmt.Sprintf("/%s%s", apiVersion, "/system/backup"),
		grpcx.NewJSONBodyHTTPHandle(backupFactory, postBackupHandler))
}

func postBackupHandler(value interface{}) (*grpcx.JSONResult, error) {
	err := Store.BackupTo(value.(*backup).ToAddr)
	if err != nil {
		log.Errorf("api-backup-post: errors:%+v", err)
		return nil, err
	}

	return &grpcx.JSONResult{}, nil
}

func backupFactory() interface{} {
	return &backup{}
}

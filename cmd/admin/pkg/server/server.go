package server

import (
	"sync"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/task"
	"github.com/labstack/echo"
	sd "github.com/labstack/echo/engine/standard"
	mw "github.com/labstack/echo/middleware"
)

// Result http interface return json
type Result struct {
	Code  int         `json:"code, omitempty"`
	Error string      `json:"error"`
	Value interface{} `json:"value"`
}

// AdminServer http interface server
type AdminServer struct {
	user       string
	pwd        string
	addr       string
	e          *echo.Echo
	store      model.Store
	taskRunner *task.Runner

	stopped  int32
	stopC    chan struct{}
	stopOnce sync.Once
	stopWG   sync.WaitGroup
}

// NewAdminServer create a AdminServer
func NewAdminServer(addr string, registryAddr string, prefix string, user string, pwd string) *AdminServer {
	taskRunner := task.NewRunner()
	st, _ := model.GetStoreFrom(registryAddr, prefix, taskRunner)

	server := &AdminServer{
		user:  user,
		pwd:   pwd,
		e:     echo.New(),
		addr:  addr,
		store: st,
	}

	server.initHTTPServer()

	return server
}

func (server *AdminServer) initHTTPServer() {
	server.e.Use(mw.Logger())
	server.e.Use(mw.Recover())
	server.e.Use(mw.Gzip())
	server.e.Use(mw.BasicAuth(func(inputUser string, inputPwd string) bool {
		if inputUser == server.user && server.pwd == inputPwd {
			return true
		}
		return false
	}))

	server.e.Static("/assets", "public/assets")
	server.e.Static("/html", "public/html") // angular html template

	server.e.File("/", "public/html/base.html")

	server.initAPIRoute()
}

// Start start the admin server
func (server *AdminServer) Start() {
	go server.listenToStop()

	log.Infof("bootstrap: gateway admin start at <%s>", server.addr)
	httpSvr := sd.New(server.addr)
	server.e.Run(httpSvr)
}

func (server *AdminServer) listenToStop() {
	<-server.stopC
	server.doStop()
}

func (server *AdminServer) doStop() {
	server.stopOnce.Do(func() {
		defer server.stopWG.Done()
		server.taskRunner.Stop()
	})
}

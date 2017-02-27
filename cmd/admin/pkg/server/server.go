package server

import (
	"fmt"

	"github.com/fagongzi/gateway/pkg/model"
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
	user  string
	pwd   string
	addr  string
	e     *echo.Echo
	store model.Store
}

// NewAdminServer create a AdminServer
func NewAdminServer(addr string, etcdAddrs []string, etcdPrefix string, user string, pwd string) *AdminServer {
	st, _ := model.NewEtcdStore(etcdAddrs, etcdPrefix)

	st.GC()

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
	fmt.Printf("start at %s\n", server.addr)
	server.e.Run(sd.New(server.addr))
}

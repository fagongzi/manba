package server

import (
	"fmt"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/labstack/echo"
	sd "github.com/labstack/echo/engine/standard"
	mw "github.com/labstack/echo/middleware"
)

type Result struct {
	Code  int         `json:"code, omitempty"`
	Error string      `json:"error"`
	Value interface{} `json:"value"`
}

type AdminServer struct {
	user  string
	pwd   string
	addr  string
	e     *echo.Echo
	store model.Store
}

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

func (self *AdminServer) initHTTPServer() {
	self.e.Use(mw.Logger())
	self.e.Use(mw.Recover())
	self.e.Use(mw.Gzip())
	self.e.Use(mw.BasicAuth(func(inputUser string, inputPwd string) bool {
		if inputUser == self.user && self.pwd == inputPwd {
			return true
		}
		return false
	}))

	self.e.Static("/assets", "public/assets")
	self.e.Static("/html", "public/html") // angular html template

	self.e.File("/", "public/html/base.html")

	self.initApiRoute()
}

func (self *AdminServer) Start() {
	fmt.Printf("start at %s\n", self.addr)
	self.e.Run(sd.New(self.addr))
}

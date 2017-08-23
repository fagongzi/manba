package api

import (
	"sync"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/task"
	"github.com/labstack/echo"
	sd "github.com/labstack/echo/engine/standard"
	mw "github.com/labstack/echo/middleware"
)

// Server is the api server of gateway
type Server struct {
	cfg        *Cfg
	store      model.Store
	taskRunner *task.Runner
	api        *echo.Echo

	stopped  int32
	stopC    chan struct{}
	stopOnce sync.Once
	stopWG   sync.WaitGroup
}

// NewServer returns the api server
func NewServer(cfg *Cfg) (*Server, error) {
	taskRunner := task.NewRunner()
	store, err := model.GetStoreFrom(cfg.RegistryAddr, cfg.Prefix, taskRunner)
	if err != nil {
		return nil, err
	}

	server := &Server{
		cfg:        cfg,
		taskRunner: taskRunner,
		store:      store,
		api:        echo.New(),
	}

	server.initHTTPServer()

	return server, nil
}

// Start start the admin server
func (s *Server) Start() {
	go s.listenToStop()

	log.Infof("bootstrap: api server start at <%s>", s.cfg.Addr)
	httpSvr := sd.New(s.cfg.Addr)
	s.api.Run(httpSvr)
}

// Stop stop the admin
func (s *Server) Stop() {
	log.Infof("stop: start to stop api server")

	s.stopWG.Add(1)
	s.stopC <- struct{}{}
	s.stopWG.Wait()

	log.Infof("stop: api server stopped")
}

func (s *Server) listenToStop() {
	<-s.stopC
	s.doStop()
}

func (s *Server) doStop() {
	s.stopOnce.Do(func() {
		defer s.stopWG.Done()
		s.taskRunner.Stop()
	})
}

func (s *Server) initHTTPServer() {
	s.api.Use(mw.Logger())
	s.api.Use(mw.Recover())
	s.api.Use(mw.CORS())
	// s.api.Use(mw.BasicAuth(func(inputUser string, inputPwd string) bool {
	// 	if inputUser == s.cfg.User && s.cfg.Pwd == inputPwd {
	// 		return true
	// 	}
	// 	return false
	// }))

	s.initAPI()
}

func (s *Server) initAPI() {
	s.initAPIOfLBS()
	s.initAPIOfProxies()
	s.initAPIOfClusters()
	s.initAPIOfServers()
	s.initAPIOfAnalysis()
	s.initAPIOfBinds()
	s.initAPIOfAPIs()
	s.initAPIOfRoutings()
}

package api

import (
	"io"
	"io/ioutil"
	"sync"

	"github.com/fagongzi/gateway/pkg/store"
	"github.com/fagongzi/log"
	fjson "github.com/fagongzi/util/json"
	"github.com/fagongzi/util/task"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
)

// Server is the api server of gateway
type Server struct {
	cfg        *Cfg
	store      store.Store
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
	store, err := store.GetStoreFrom(cfg.RegistryAddr, cfg.Prefix, taskRunner)
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
	s.api.Start(s.cfg.Addr)
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

func readJSONFromReader(value interface{}, r io.ReadCloser) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return fjson.Unmarshal(value, data)
}

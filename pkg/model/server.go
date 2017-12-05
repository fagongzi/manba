package model

import (
	"time"

	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/log"
	validation "github.com/go-ozzo/ozzo-validation"
)



// Validate validate the model
func (s *Server) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Addr, validation.Required))
}

// HeathCheck heath check
type HeathCheck struct {
	Path     string        `json:"path,omitempty"`
	Body     string        `json:"body"`
	Interval time.Duration `json:"interval,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty"`
}

// CircuitBreaker circuit breaker
type CircuitBreaker struct {
	CloseToHalf     time.Duration `json:"closeToHalf,omitempty"`
	HalfToOpen      time.Duration `json:"halfToOpen,omitempty"`
	OpenToClose     time.Duration `json:"openToClose,omitempty"`
	HalfTrafficRate int           `json:"halfTrafficRate,omitempty"`
	HalfToOpenRate  int           `json:"halfToOpenRate,omitempty"`
	OpenToCloseRate int           `json:"openToCloseRate,omitempty"`
}

// Server server
type Server struct {
	// meta
	ID             string          `json:"id, omitempty"`
	Addr           string          `json:"addr,omitempty"`
	Schema         string          `json:"schema,omitempty"`
	MaxQPS         int             `json:"maxQPS,omitempty"`
	HeathCheck     *HeathCheck     `json:"heathCheck, omitempty"`
	CircuitBreaker *CircuitBreaker `json:"circuitBreaker, omitempty"`
	External       bool            `json:"external,omitempty"`
	BindClusters   []string        `json:"bindClusters,omitempty"`
}

// Init init model
func (s *Server) Init() error {
	if s.ID == "" {
		s.ID = util.NewID()
	}

	return nil
}

// HasBind add bind
func (s *Server) HasBind() bool {
	return len(s.BindClusters) > 0
}

// AddBind add bind
func (s *Server) AddBind(bind *Bind) {
	index := s.indexOf(bind.ClusterID)
	if index == -1 {
		s.BindClusters = append(s.BindClusters, bind.ClusterID)
	}
}

// RemoveBind remove bind
func (s *Server) RemoveBind(id string) {
	index := s.indexOf(id)
	if index >= 0 {
		s.BindClusters = append(s.BindClusters[:index], s.BindClusters[index+1:]...)
	}
}

func (s *Server) indexOf(id string) int {
	for index, s := range s.BindClusters {
		if s == id {
			return index
		}
	}

	return -1
}

func (s *Server) updateFrom(svr *Server) {
	s.External = svr.External
	s.Schema = svr.Schema
	s.Addr = svr.Addr
	s.MaxQPS = svr.MaxQPS
	s.HeathCheck = svr.HeathCheck
	s.CircuitBreaker = svr.CircuitBreaker
	s.BindClusters = svr.BindClusters

	log.Infof("meta: server <%s> updated",
		s.Addr)
}

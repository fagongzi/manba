package model

import (
	"time"

	"github.com/fagongzi/util/uuid"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Validate validate the model
func (s *Server) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Addr, validation.Required),
		validation.Field(&s.Schema, validation.Required),
		validation.Field(&s.MaxQPS, validation.Required))
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
	CloseTimeout       time.Duration `json:"closeTimeout,omitempty"`
	HalfTrafficRate    int           `json:"halfTrafficRate,omitempty"`
	RateCheckPeriod    time.Duration `json:"rateCheckPeriod,omitempty"`
	FailureRateToClose int           `json:"failureRateToClose,omitempty"`
	SucceedRateToOpen  int           `json:"succeedRateToOpen,omitempty"`
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
}

// Init init model
func (s *Server) Init() error {
	if s.ID == "" {
		s.ID = uuid.NewID()
	}

	return nil
}

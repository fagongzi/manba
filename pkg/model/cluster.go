package model

import (
	"github.com/fagongzi/gateway/pkg/lb"
	"github.com/fagongzi/util/uuid"
	"github.com/go-ozzo/ozzo-validation"
)

// Cluster cluster
type Cluster struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	LbName string `json:"lbName,omitempty"`
}

// Validate validate the model
func (c *Cluster) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.LbName, validation.Required, validation.In(lb.ROUNDROBIN)))
}

// Init init
func (c *Cluster) Init() error {
	if c.ID == "" {
		c.ID = uuid.NewID()
	}

	return nil
}

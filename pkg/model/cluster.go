package model

import (
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/go-ozzo/ozzo-validation"
)

// Cluster cluster
type Cluster struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	LbName      string   `json:"lbName,omitempty"`
	BindServers []string `json:"bindServers,omitempty"`
}

// Validate validate the model
func (c *Cluster) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.LbName, validation.Required))
}

// AddBind add bind
func (c *Cluster) AddBind(bind *Bind) {
	index := c.indexOf(bind.ServerID)
	if index == -1 {
		c.BindServers = append(c.BindServers, bind.ServerID)
	}
}

// HasBind add bind
func (c *Cluster) HasBind() bool {
	return len(c.BindServers) > 0
}

// RemoveBind remove bind
func (c *Cluster) RemoveBind(id string) {
	index := c.indexOf(id)
	if index >= 0 {
		c.BindServers = append(c.BindServers[:index], c.BindServers[index+1:]...)
	}
}

func (c *Cluster) indexOf(id string) int {
	for index, s := range c.BindServers {
		if s == id {
			return index
		}
	}

	return -1
}

// Init init
func (c *Cluster) Init() error {
	if c.ID == "" {
		c.ID = util.NewID()
	}

	return nil
}

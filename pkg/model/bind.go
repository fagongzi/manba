package model

import (
	"github.com/fagongzi/gateway/pkg/util"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Bind a bind server and cluster
type Bind struct {
	ID        string `json:"id, omitempty"`
	ClusterID string `json:"clusterID, omitempty"`
	ServerID  string `json:"serverID, omitempty"`
}

// Validate validate the model
func (b *Bind) Validate() error {
	return validation.ValidateStruct(b,
		validation.Field(&b.ClusterID, validation.Required),
		validation.Field(&b.ServerID, validation.Required))
}

// Init init model
func (b *Bind) Init() error {
	if b.ID == "" {
		b.ID = util.NewID()
	}

	return nil
}

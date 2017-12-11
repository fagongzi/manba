package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

// Bind a bind server and cluster
type Bind struct {
	ClusterID string `json:"clusterID, omitempty"`
	ServerID  string `json:"serverID, omitempty"`
}

// Validate validate the model
func (b *Bind) Validate() error {
	return validation.ValidateStruct(b,
		validation.Field(&b.ClusterID, validation.Required),
		validation.Field(&b.ServerID, validation.Required))
}

package client

import (
	"github.com/fagongzi/gateway/pkg/pb"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
)

// PluginBuilder plugin builder
type PluginBuilder struct {
	c     *client
	value metapb.Plugin
}

// NewPluginBuilder return a plugin build
func (c *client) NewPluginBuilder() *PluginBuilder {
	return &PluginBuilder{
		c: c,
		value: metapb.Plugin{
			Type: metapb.JavaScript,
		},
	}
}

// Use use a plugin
func (sb *PluginBuilder) Use(value metapb.Plugin) *PluginBuilder {
	sb.value = value
	return sb
}

// Name set plugin name
func (sb *PluginBuilder) Name(name string) *PluginBuilder {
	sb.value.Name = name
	return sb
}

// Version set plugin version
func (sb *PluginBuilder) Version(version int64) *PluginBuilder {
	sb.value.Version = version
	return sb
}

// Author set plugin author
func (sb *PluginBuilder) Author(author, email string) *PluginBuilder {
	sb.value.Author = author
	sb.value.Email = email
	return sb
}

// Script set plugin script
func (sb *PluginBuilder) Script(content, cfg []byte) *PluginBuilder {
	sb.value.Content = content
	sb.value.Cfg = cfg
	return sb
}

// Commit commit
func (sb *PluginBuilder) Commit() (uint64, error) {
	err := pb.ValidatePlugin(&sb.value)
	if err != nil {
		return 0, err
	}

	return sb.c.putPlugin(sb.value)
}

// Build build
func (sb *PluginBuilder) Build() (*rpcpb.PutPluginReq, error) {
	err := pb.ValidatePlugin(&sb.value)
	if err != nil {
		return nil, err
	}

	return &rpcpb.PutPluginReq{
		Plugin: sb.value,
	}, nil
}

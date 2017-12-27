package client

import (
	"github.com/fagongzi/gateway/pkg/pb"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

// APIBuilder api builder
type APIBuilder struct {
	c     *client
	value metapb.API
}

// NewAPIBuilder return a api build
func (c *client) NewAPIBuilder() *APIBuilder {
	return &APIBuilder{
		c:     c,
		value: metapb.API{},
	}
}

// Use use a cluster
func (ab *APIBuilder) Use(value metapb.API) *APIBuilder {
	ab.value = value
	return ab
}

// Name set a name
func (ab *APIBuilder) Name(name string) *APIBuilder {
	ab.value.Name = name
	return ab
}

// AuthPlugin set a auth filter plugin
func (ab *APIBuilder) AuthPlugin(name string) *APIBuilder {
	ab.value.AuthFilter = name
	return ab
}

// AddPerm add a perm
func (ab *APIBuilder) AddPerm(perm string) *APIBuilder {
	ab.value.Perms = append(ab.value.Perms, perm)
	return ab
}

// RemovePerm remove a perm
func (ab *APIBuilder) RemovePerm(perm string) *APIBuilder {
	if len(ab.value.Perms) == 0 {
		return ab
	}

	var perms []string
	for _, p := range ab.value.Perms {
		if p != perm {
			perms = append(perms, p)
		}
	}

	ab.value.Perms = perms
	return ab
}

// MatchURLPattern set a match path
func (ab *APIBuilder) MatchURLPattern(urlPattern string) *APIBuilder {
	ab.value.URLPattern = urlPattern
	ab.value.Domain = ""
	return ab
}

// MatchMethod set a match method
func (ab *APIBuilder) MatchMethod(method string) *APIBuilder {
	ab.value.Method = method
	ab.value.Domain = ""
	return ab
}

// MatchDomain set a match domain
func (ab *APIBuilder) MatchDomain(domain string) *APIBuilder {
	ab.value.Domain = domain
	ab.value.Method = ""
	ab.value.URLPattern = ""
	return ab
}

// UP up this api
func (ab *APIBuilder) UP() *APIBuilder {
	ab.value.Status = metapb.Up
	return ab
}

// Down down this api
func (ab *APIBuilder) Down() *APIBuilder {
	ab.value.Status = metapb.Down
	return ab
}

// NoDefaultValue set no default value
func (ab *APIBuilder) NoDefaultValue() *APIBuilder {
	ab.value.DefaultValue = nil
	return ab
}

// DefaultValue set default value
func (ab *APIBuilder) DefaultValue(value []byte) *APIBuilder {
	if ab.value.DefaultValue == nil {
		ab.value.DefaultValue = &metapb.HTTPResult{}
	}

	ab.value.DefaultValue.Body = value
	return ab
}

// AddDefaultValueHeader add default value header
func (ab *APIBuilder) AddDefaultValueHeader(name, value string) *APIBuilder {
	if ab.value.DefaultValue == nil {
		ab.value.DefaultValue = &metapb.HTTPResult{}
	}

	ab.value.DefaultValue.Headers = append(ab.value.DefaultValue.Headers, &metapb.PairValue{
		Name:  name,
		Value: value,
	})
	return ab
}

// AddDefaultValueCookie add default value cookie
func (ab *APIBuilder) AddDefaultValueCookie(name, value string) *APIBuilder {
	if ab.value.DefaultValue == nil {
		ab.value.DefaultValue = &metapb.HTTPResult{}
	}

	ab.value.DefaultValue.Cookies = append(ab.value.DefaultValue.Cookies, &metapb.PairValue{
		Name:  name,
		Value: value,
	})
	return ab
}

// NoWhitelist set no whiltelist
func (ab *APIBuilder) NoWhitelist() *APIBuilder {
	if ab.value.IPAccessControl == nil {
		return ab
	}

	ab.value.IPAccessControl.Whitelist = make([]string, 0, 0)
	return ab
}

// NoBlacklist set no blacklist
func (ab *APIBuilder) NoBlacklist() *APIBuilder {
	if ab.value.IPAccessControl == nil {
		return ab
	}

	ab.value.IPAccessControl.Blacklist = make([]string, 0, 0)
	return ab
}

// RemoveWhitelist remove ip white list
func (ab *APIBuilder) RemoveWhitelist(ips ...string) *APIBuilder {
	if ab.value.IPAccessControl == nil || len(ab.value.IPAccessControl.Whitelist) == 0 {
		return ab
	}

	var value []string
	for _, old := range ab.value.IPAccessControl.Whitelist {
		for _, ip := range ips {
			if old != ip {
				value = append(value, old)
			}
		}
	}

	ab.value.IPAccessControl.Whitelist = value
	return ab
}

// RemoveBlacklist remove ip white list
func (ab *APIBuilder) RemoveBlacklist(ips ...string) *APIBuilder {
	if ab.value.IPAccessControl == nil || len(ab.value.IPAccessControl.Blacklist) == 0 {
		return ab
	}

	var value []string
	for _, old := range ab.value.IPAccessControl.Blacklist {
		for _, ip := range ips {
			if old != ip {
				value = append(value, old)
			}
		}
	}

	ab.value.IPAccessControl.Blacklist = value
	return ab
}

// AddWhitelist add ip white list
func (ab *APIBuilder) AddWhitelist(ips ...string) *APIBuilder {
	if ab.value.IPAccessControl == nil {
		ab.value.IPAccessControl = &metapb.IPAccessControl{}
	}

	ab.value.IPAccessControl.Whitelist = append(ab.value.IPAccessControl.Whitelist, ips...)
	return ab
}

// AddBlacklist add ip black list
func (ab *APIBuilder) AddBlacklist(ips ...string) *APIBuilder {
	if ab.value.IPAccessControl == nil {
		ab.value.IPAccessControl = &metapb.IPAccessControl{}
	}

	ab.value.IPAccessControl.Blacklist = append(ab.value.IPAccessControl.Blacklist, ips...)
	return ab
}

// AddDispatchNode add a dispatch node
func (ab *APIBuilder) AddDispatchNode(cluster uint64) *APIBuilder {
	for _, n := range ab.value.Nodes {
		if n.ClusterID == cluster {
			return ab
		}
	}

	ab.value.Nodes = append(ab.value.Nodes, &metapb.DispatchNode{
		ClusterID: cluster,
	})

	return ab
}

// RemoveDispatchNodeURLRewrite remove dispatch node
func (ab *APIBuilder) RemoveDispatchNodeURLRewrite(cluster uint64) *APIBuilder {
	var nodes []*metapb.DispatchNode

	for _, n := range ab.value.Nodes {
		if n.ClusterID != cluster {
			nodes = append(nodes, n)
		}
	}

	ab.value.Nodes = nodes
	return ab
}

// DispatchNodeURLRewrite set dispatch node url rewrite
func (ab *APIBuilder) DispatchNodeURLRewrite(cluster uint64, urlRewrite string) *APIBuilder {
	var node *metapb.DispatchNode

	for _, n := range ab.value.Nodes {
		if n.ClusterID == cluster {
			node = n
			break
		}
	}

	if node == nil {
		ab.value.Nodes = append(ab.value.Nodes, &metapb.DispatchNode{
			ClusterID:  cluster,
			URLRewrite: urlRewrite,
		})
	} else {
		node.URLRewrite = urlRewrite
	}

	return ab
}

// DispatchNodeValueAttrName set dispatch node attr name of value
func (ab *APIBuilder) DispatchNodeValueAttrName(cluster uint64, attrName string) *APIBuilder {
	var node *metapb.DispatchNode

	for _, n := range ab.value.Nodes {
		if n.ClusterID == cluster {
			node = n
			break
		}
	}

	if node == nil {
		ab.value.Nodes = append(ab.value.Nodes, &metapb.DispatchNode{
			ClusterID: cluster,
			AttrName:  attrName,
		})
	} else {
		node.AttrName = attrName
	}

	return ab
}

// AddDispatchNodeValidation add dispatch node validation
func (ab *APIBuilder) AddDispatchNodeValidation(cluster uint64, param metapb.Parameter, rule string, required bool) *APIBuilder {
	var node *metapb.DispatchNode

	for _, n := range ab.value.Nodes {
		if n.ClusterID == cluster {
			node = n
			break
		}
	}

	if node == nil {
		node = &metapb.DispatchNode{
			ClusterID: cluster,
		}
		ab.value.Nodes = append(ab.value.Nodes, node)
	}

	var validation *metapb.Validation
	for _, v := range node.Validations {
		if v.Parameter.Name == param.Name && v.Parameter.Source == param.Source {
			validation = v
			break
		}
	}

	if validation == nil {
		validation = &metapb.Validation{
			Parameter: param,
			Required:  required,
		}
		node.Validations = append(node.Validations, validation)
	}

	validation.Rules = append(validation.Rules, metapb.ValidationRule{
		RuleType:   metapb.RuleRegexp,
		Expression: rule,
	})

	return ab
}

// Commit commit
func (ab *APIBuilder) Commit() (uint64, error) {
	err := pb.ValidateAPI(&ab.value)
	if err != nil {
		return 0, err
	}

	return ab.c.putAPI(ab.value)
}

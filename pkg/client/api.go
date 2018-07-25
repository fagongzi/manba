package client

import (
	"time"

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

// UseDefaultValue use default value if force
func (ab *APIBuilder) UseDefaultValue(force bool) *APIBuilder {
	ab.value.UseDefault = force
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

// AppendDispatchNode append a dispatch node even if the cluster added
func (ab *APIBuilder) AppendDispatchNode(cluster uint64) *APIBuilder {
	ab.value.Nodes = append(ab.value.Nodes, &metapb.DispatchNode{
		ClusterID: cluster,
	})

	return ab
}

// AddDispatchNode add a dispatch node if the cluster not added
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

// DispatchNodeBatchIndex add a dispatch node batch index
func (ab *APIBuilder) DispatchNodeBatchIndex(cluster uint64, batchIndex int) *APIBuilder {
	return ab.DispatchNodeBatchIndexWithIndex(cluster, 0, batchIndex)
}

// DispatchNodeBatchIndexWithIndex add a dispatch node batch index
func (ab *APIBuilder) DispatchNodeBatchIndexWithIndex(cluster uint64, idx int, batchIndex int) *APIBuilder {
	node := ab.getNode(cluster, idx)
	if nil == node {
		ab.value.Nodes = append(ab.value.Nodes, &metapb.DispatchNode{
			ClusterID:  cluster,
			BatchIndex: int32(batchIndex),
		})
	} else {
		node.BatchIndex = int32(batchIndex)
	}

	return ab
}

// AddDispatchNodeDefaultValue add default value for dispatch
func (ab *APIBuilder) AddDispatchNodeDefaultValue(cluster uint64, value []byte) *APIBuilder {
	return ab.AddDispatchNodeDefaultValueWithIndex(cluster, 0, value)
}

// AddDispatchNodeDefaultValueWithIndex add default value for dispatch
func (ab *APIBuilder) AddDispatchNodeDefaultValueWithIndex(cluster uint64, idx int, value []byte) *APIBuilder {
	node := ab.getNode(cluster, idx)
	if nil == node {
		ab.value.Nodes = append(ab.value.Nodes, &metapb.DispatchNode{
			ClusterID: cluster,
			DefaultValue: &metapb.HTTPResult{
				Body: value,
			},
		})
	} else {
		if node.DefaultValue == nil {
			node.DefaultValue = &metapb.HTTPResult{
				Body: value,
			}
		} else {
			node.DefaultValue.Body = value
		}
	}

	return ab
}

// UseDispatchNodeDefaultValue use default value if force
func (ab *APIBuilder) UseDispatchNodeDefaultValue(cluster uint64, force bool) *APIBuilder {
	return ab.UseDispatchNodeDefaultValueWithIndex(cluster, 0, force)
}

// UseDispatchNodeDefaultValueWithIndex use default value if force
func (ab *APIBuilder) UseDispatchNodeDefaultValueWithIndex(cluster uint64, idx int, force bool) *APIBuilder {
	node := ab.getNode(cluster, idx)
	if nil == node {
		ab.value.Nodes = append(ab.value.Nodes, &metapb.DispatchNode{
			ClusterID:  cluster,
			UseDefault: force,
		})
	} else {
		node.UseDefault = force
	}

	return ab
}

// AddDispatchNodeDefaultValueHeader add default value header
func (ab *APIBuilder) AddDispatchNodeDefaultValueHeader(cluster uint64, name, value string) *APIBuilder {
	return ab.AddDispatchNodeDefaultValueHeaderWithIndex(cluster, 0, name, value)
}

// AddDispatchNodeDefaultValueHeaderWithIndex add default value header
func (ab *APIBuilder) AddDispatchNodeDefaultValueHeaderWithIndex(cluster uint64, idx int, name, value string) *APIBuilder {
	node := ab.getNode(cluster, idx)
	if nil == node {
		ab.value.Nodes = append(ab.value.Nodes, &metapb.DispatchNode{
			ClusterID: cluster,
			DefaultValue: &metapb.HTTPResult{
				Headers: []*metapb.PairValue{
					&metapb.PairValue{
						Name:  name,
						Value: value,
					},
				},
			},
		})
	} else {
		if node.DefaultValue == nil {
			node.DefaultValue = &metapb.HTTPResult{
				Headers: []*metapb.PairValue{
					&metapb.PairValue{
						Name:  name,
						Value: value,
					},
				},
			}
		} else {
			node.DefaultValue.Headers = append(node.DefaultValue.Headers, &metapb.PairValue{
				Name:  name,
				Value: value,
			})
		}
	}

	return ab
}

// AddDispatchNodeDefaultValueCookie add default value cookie
func (ab *APIBuilder) AddDispatchNodeDefaultValueCookie(cluster uint64, name, value string) *APIBuilder {
	return ab.AddDispatchNodeDefaultValueCookieWithIndex(cluster, 0, name, value)
}

// AddDispatchNodeDefaultValueCookieWithIndex add default value cookie
func (ab *APIBuilder) AddDispatchNodeDefaultValueCookieWithIndex(cluster uint64, idx int, name, value string) *APIBuilder {
	node := ab.getNode(cluster, idx)
	if nil == node {
		ab.value.Nodes = append(ab.value.Nodes, &metapb.DispatchNode{
			ClusterID: cluster,
			DefaultValue: &metapb.HTTPResult{
				Cookies: []*metapb.PairValue{
					&metapb.PairValue{
						Name:  name,
						Value: value,
					},
				},
			},
		})
	} else {
		if node.DefaultValue == nil {
			node.DefaultValue = &metapb.HTTPResult{
				Cookies: []*metapb.PairValue{
					&metapb.PairValue{
						Name:  name,
						Value: value,
					},
				},
			}
		} else {
			node.DefaultValue.Cookies = append(node.DefaultValue.Cookies, &metapb.PairValue{
				Name:  name,
				Value: value,
			})
		}
	}

	return ab
}

// RemoveDispatchNodeURLRewrite remove dispatch node
func (ab *APIBuilder) RemoveDispatchNodeURLRewrite(cluster uint64) *APIBuilder {
	for _, n := range ab.value.Nodes {
		if n.ClusterID == cluster {
			n.URLRewrite = ""
		}
	}
	return ab
}

// DispatchNodeUseCachingWithIndex set dispatch node caching
func (ab *APIBuilder) DispatchNodeUseCachingWithIndex(cluster uint64, index int, deadline time.Duration) *APIBuilder {
	node := ab.getNode(cluster, index)

	if node == nil {
		ab.value.Nodes = append(ab.value.Nodes, &metapb.DispatchNode{
			ClusterID: cluster,
			Cache: &metapb.Cache{
				Deadline: uint64(deadline.Seconds()),
			},
		})
	} else {
		node.Cache = &metapb.Cache{
			Deadline: uint64(deadline.Seconds()),
		}
	}

	return ab
}

// DispatchNodeUseCaching set dispatch node caching
func (ab *APIBuilder) DispatchNodeUseCaching(cluster uint64, deadline time.Duration) *APIBuilder {
	return ab.DispatchNodeUseCachingWithIndex(cluster, 0, deadline)
}

// AddDispatchNodeCachingKeyWithIndex add key for caching
func (ab *APIBuilder) AddDispatchNodeCachingKeyWithIndex(cluster uint64, index int, keys ...metapb.Parameter) *APIBuilder {
	node := ab.getNode(cluster, index)
	if node != nil {
		node.Cache.Keys = append(node.Cache.Keys, keys...)
	}

	return ab
}

// AddDispatchNodeCachingKey add key for caching
func (ab *APIBuilder) AddDispatchNodeCachingKey(cluster uint64, keys ...metapb.Parameter) *APIBuilder {
	return ab.AddDispatchNodeCachingKeyWithIndex(cluster, 0, keys...)
}

// AddDispatchNodeCachingConditionWithIndex add condition for caching
func (ab *APIBuilder) AddDispatchNodeCachingConditionWithIndex(cluster uint64, index int, param metapb.Parameter, op metapb.CMP, expect string) *APIBuilder {
	node := ab.getNode(cluster, index)
	if node != nil {
		node.Cache.Conditions = append(node.Cache.Conditions, metapb.Condition{
			Parameter: param,
			Cmp:       op,
			Expect:    expect,
		})
	}

	return ab
}

// AddDispatchNodeCachingCondition add condition for caching
func (ab *APIBuilder) AddDispatchNodeCachingCondition(cluster uint64, param metapb.Parameter, op metapb.CMP, expect string) *APIBuilder {
	return ab.AddDispatchNodeCachingConditionWithIndex(cluster, 0, param, op, expect)
}

// DispatchNodeURLRewriteWithIndex set dispatch node url rewrite
func (ab *APIBuilder) DispatchNodeURLRewriteWithIndex(cluster uint64, index int, urlRewrite string) *APIBuilder {
	node := ab.getNode(cluster, index)

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

// DispatchNodeURLRewrite set dispatch node url rewrite
func (ab *APIBuilder) DispatchNodeURLRewrite(cluster uint64, urlRewrite string) *APIBuilder {
	return ab.DispatchNodeURLRewriteWithIndex(cluster, 0, urlRewrite)
}

// DispatchNodeValueAttrNameWithIndex set dispatch node attr name of value
func (ab *APIBuilder) DispatchNodeValueAttrNameWithIndex(cluster uint64, index int, attrName string) *APIBuilder {
	node := ab.getNode(cluster, index)

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

// DispatchNodeValueAttrName set dispatch node attr name of value
func (ab *APIBuilder) DispatchNodeValueAttrName(cluster uint64, attrName string) *APIBuilder {
	return ab.DispatchNodeValueAttrNameWithIndex(cluster, 0, attrName)
}

// AddDispatchNodeValidationWithIndex add dispatch node validation
func (ab *APIBuilder) AddDispatchNodeValidationWithIndex(cluster uint64, index int, param metapb.Parameter, rule string, required bool) *APIBuilder {
	node := ab.getNode(cluster, index)

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

// AddDispatchNodeValidation add dispatch node validation
func (ab *APIBuilder) AddDispatchNodeValidation(cluster uint64, param metapb.Parameter, rule string, required bool) *APIBuilder {
	return ab.AddDispatchNodeValidationWithIndex(cluster, 0, param, rule, required)
}

// NoRenderTemplate clear render template
func (ab *APIBuilder) NoRenderTemplate() *APIBuilder {
	ab.value.RenderTemplate = nil
	return ab
}

// AddFlatRenderObject add the render object to the top level object
func (ab *APIBuilder) AddFlatRenderObject(namesAndExtractExps ...string) *APIBuilder {
	return ab.addRenderObject("", true, namesAndExtractExps...)
}

// AddRenderObject add the render object to the top level object
func (ab *APIBuilder) AddRenderObject(nameInTemplate string, namesAndExtractExps ...string) *APIBuilder {
	return ab.addRenderObject(nameInTemplate, false, namesAndExtractExps...)
}

func (ab *APIBuilder) addRenderObject(nameInTemplate string, flatAttrs bool, namesAndExtractExps ...string) *APIBuilder {
	if len(namesAndExtractExps) == 0 || len(namesAndExtractExps)%2 != 0 {
		return ab
	}

	if ab.value.RenderTemplate == nil {
		ab.value.RenderTemplate = &metapb.RenderTemplate{}
	}

	var obj *metapb.RenderObject
	for _, o := range ab.value.RenderTemplate.Objects {
		if o.Name == nameInTemplate {
			obj = o
		}
	}

	if obj == nil {
		obj = &metapb.RenderObject{
			Name:      nameInTemplate,
			FlatAttrs: flatAttrs,
		}
		ab.value.RenderTemplate.Objects = append(ab.value.RenderTemplate.Objects, obj)
	}

	l := len(namesAndExtractExps) / 2
	for i := 0; i < l; i++ {
		obj.Attrs = append(obj.Attrs, &metapb.RenderAttr{
			Name:       namesAndExtractExps[2*i],
			ExtractExp: namesAndExtractExps[2*i+1],
		})
	}

	return ab
}

// AddTag add tag for api
func (ab *APIBuilder) AddTag(key, value string) *APIBuilder {
	ab.value.Tags = append(ab.value.Tags, &metapb.PairValue{
		Name:  key,
		Value: value,
	})
	return ab
}

// RemoveTag remove tag for api
func (ab *APIBuilder) RemoveTag(key string) *APIBuilder {
	var newTags []*metapb.PairValue
	for _, tag := range ab.value.Tags {
		if tag.Name != key {
			newTags = append(newTags, tag)
		}
	}
	ab.value.Tags = newTags
	return ab
}

// Position reset the position for api
func (ab *APIBuilder) Position(value uint32) *APIBuilder {
	ab.value.Position = value
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

func (ab *APIBuilder) getNode(cluster uint64, index int) *metapb.DispatchNode {
	var node *metapb.DispatchNode

	idx := 0
	for _, n := range ab.value.Nodes {
		if n.ClusterID == cluster && idx == index {
			node = n
			break
		}

		if n.ClusterID == cluster {
			idx++
		}
	}

	return node
}

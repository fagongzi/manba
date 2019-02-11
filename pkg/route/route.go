package route

import (
	"bytes"
	"fmt"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/hack"
)

type routeItem struct {
	node     node
	children []*routeItem
	api      uint64
}

func (item *routeItem) removeAPI(api uint64) bool {
	if item.api == api {
		item.api = 0
		return true
	}

	for _, c := range item.children {
		if c.removeAPI(api) {
			return true
		}
	}

	return false
}

func (item *routeItem) addChildren(id uint64, nodes ...node) {
	parent := item

	for _, n := range nodes {
		p := &routeItem{
			node: n,
		}
		parent.children = append(parent.children, p)
		parent = p
	}

	parent.api = id
}

func (item *routeItem) matches(n node) bool {
	if item.node.nt != n.nt {
		return false
	}

	switch item.node.nt {
	case slashType:
		return true
	case numberType:
		return true
	case stringType:
		return true
	case constType:
		return bytes.Compare(item.node.value, n.value) == 0
	case enumType:
		return true
	default:
		log.Fatalf("bug: error node type %d", item.node.nt)
	}

	return false
}

// Route route for api match
// url define: /conststring/(number|string|enum:m1|m2|m3)[:argname]
type Route struct {
	root *routeItem
}

// NewRoute returns a route
func NewRoute() *Route {
	return &Route{
		root: &routeItem{
			node: node{
				nt: slashType,
			},
		},
	}
}

// Add add a url to this route
func (r *Route) Add(api metapb.API) error {
	p := newParser(hack.StringToSlice(api.URLPattern))
	nodes, err := p.parse()
	if err != nil {
		return err
	}

	nodes = removeSlash(nodes...)
	parent := r.root
	matchedIdx := 0
	for idx, node := range nodes {
		if idx != 0 && node.nt == slashType {
			continue
		}

		if parent.matches(node) {
			matchedIdx = idx
			continue
		}

		matched := false
		for _, item := range parent.children {
			if item.matches(node) {
				parent = item
				matched = true
				matchedIdx = idx
				break
			}
		}

		if !matched {
			break
		}
	}

	if matchedIdx == len(nodes)-1 {
		if parent.api != 0 {
			return fmt.Errorf("conflict with api %d", parent.api)
		}

		parent.api = api.ID
		return nil
	}

	parent.addChildren(api.ID, nodes[matchedIdx+1:]...)
	return nil
}

// Remove remove api
func (r *Route) Remove(api uint64) bool {
	return r.root.removeAPI(api)
}

func removeSlash(nodes ...node) []node {
	var value []node

	for idx, node := range nodes {
		if node.nt != slashType {
			value = append(value, node)
		} else if node.nt == slashType && idx == 0 {
			value = append(value, node)
		}
	}

	return value
}

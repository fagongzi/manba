package route

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

type nodeType int

const (
	slashType  = nodeType(1)
	constType  = nodeType(2)
	numberType = nodeType(3)
	stringType = nodeType(4)
	enumType   = nodeType(5)
)

// Route route for api match
// url define: /conststring/(number|string|enum:m1|m2|m3)[:argname]
type Route struct {
	root *node
}

type node struct {
	nt       nodeType
	value    []byte
	enums    [][]byte
	argName  []byte
	children []node
}

func (n *node) isEnum() bool {
	return n.nt == enumType
}

func (n *node) isConst() bool {
	return n.nt == constType
}

func (n *node) addEnum(value []byte) {
	n.enums = append(n.enums, value)
}

func (n *node) setArgName(value []byte) {
	n.argName = value
}

// Add add a url to this route
func (r *Route) Add(api metapb.API) error {
	return nil
}

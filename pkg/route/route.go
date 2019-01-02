package route

type nodeType int

const (
	numberType = nodeType(1)
	stringType = nodeType(2)
	expType    = nodeType(3)
)

// Route route for api match
// url define: /path/:namedArg[type]
type Route struct {
	root *node
}

type node struct {
	value    []byte
	nt       nodeType
	children []node
}

// Add add a url to this route
func (r *Route) Add(url []byte) error {
	n := len(url)
	for i := 0; i < n; i++ {

	}
	return nil
}

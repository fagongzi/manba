package route

type tree struct {
	root *treeNode
}

func (t *tree) add(nodes []node) error {
	return nil
}

// func (t *tree) find(nodes []node) *treeNode {
// 	if len(nodes) == 1 {
// 		return t.root
// 	}
// }

type treeNode struct {
	size     int
	apis     map[nodeType]uint64
	children []treeNode
}

// func (tn *treeNode) matches(n node) bool {
// 	for i := 0; i < tn.size; i++ {

// 	}
// }

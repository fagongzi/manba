package goetty

type simpleQueue struct {
	items []interface{}
}

func (q *simpleQueue) len() int {
	return len(q.items)
}

func newSimpleQueue() *simpleQueue {
	return &simpleQueue{}
}

func (q *simpleQueue) push(item interface{}) {
	q.items = append(q.items, item)
}

func (q *simpleQueue) pop() interface{} {
	value := q.items[0]
	q.items[0] = nil
	q.items = q.items[1:]
	return value
}

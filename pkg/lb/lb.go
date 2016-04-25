package lb

import (
	"container/list"
	"net/http"
)

const (
	ROUNDROBIN = "ROUNDROBIN"
)

var (
	supportLbs = []string{ROUNDROBIN}
)

var (
	LBS = map[string]func() LoadBalance{
		ROUNDROBIN: NewRoundRobin,
	}
)

type LoadBalance interface {
	Select(req *http.Request, servers *list.List) int
}

func GetSupportLBS() []string {
	return supportLbs
}

func NewLoadBalance(name string) LoadBalance {
	return LBS[name]()
}

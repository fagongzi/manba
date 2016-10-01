package model

import (
	"container/list"
	"encoding/json"
	"io"
	"regexp"
	"sync"

	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/pkg/lb"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/valyala/fasthttp"
)

// Cluster cluster
type Cluster struct {
	Name        string   `json:"name,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	LbName      string   `json:"lbName,omitempty"`
	BindServers []string `json:"bindServers,omitempty"`

	regexp *regexp.Regexp
	svrs   *list.List
	rwLock *sync.RWMutex
	lb     lb.LoadBalance
}

// UnMarshalCluster unmarshal
func UnMarshalCluster(data []byte) *Cluster {
	v := &Cluster{}
	err := json.Unmarshal(data, v)

	if err != nil {
		return v
	}

	c, _ := NewCluster(v.Name, v.Pattern, v.LbName)

	return c
}

// UnMarshalClusterFromReader unmarshal from reader
func UnMarshalClusterFromReader(r io.Reader) (*Cluster, error) {
	v := &Cluster{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	if nil != err {
		return nil, err
	}

	return v, v.init()
}

// NewCluster create a cluster
func NewCluster(name string, pattern string, lbName string) (*Cluster, error) {
	c := &Cluster{
		Name:    name,
		Pattern: pattern,
		LbName:  lbName,
	}

	return c, c.init()
}

func (c *Cluster) init() error {
	reg, err := regexp.Compile(c.Pattern)

	if nil != err {
		return err
	}

	c.regexp = reg
	c.svrs = list.New()
	c.lb = lb.NewLoadBalance(c.LbName)
	c.rwLock = &sync.RWMutex{}

	return nil
}

func (c *Cluster) updateFrom(cluster *Cluster) {
	if c.rwLock != nil {
		c.rwLock.Lock()
		defer c.rwLock.Unlock()
	}

	c.Pattern = cluster.Pattern
	c.LbName = cluster.LbName

	c.regexp, _ = regexp.Compile(c.Pattern)
	c.lb = lb.NewLoadBalance(c.LbName)

	log.Infof("Cluster <%s> updated, %+v", c.Name, c)
}

func (c *Cluster) doInEveryBindServers(callback func(string)) {
	c.rwLock.Lock()
	defer c.rwLock.Unlock()

	for iter := c.svrs.Back(); iter != nil; iter = iter.Prev() {
		addr, _ := iter.Value.(string)
		callback(addr)
	}
}

func (c *Cluster) unbind(svr *Server) {
	c.rwLock.Lock()
	defer c.rwLock.Unlock()

	c.doUnBind(svr)
}

func (c *Cluster) doUnBind(svr *Server) {
	util.Remove(c.svrs, svr.Addr)
	log.Infof("UnBind <%s,%s> succ.", svr.Addr, c.Name)
}

func (c *Cluster) bind(svr *Server) {
	c.rwLock.Lock()
	defer c.rwLock.Unlock()

	if util.IndexOf(c.svrs, svr.Addr) >= 0 {
		log.Infof("Bind <%s,%s> already created.", svr.Addr, c.Name)
		return
	}

	c.svrs.PushBack(svr.Addr)

	log.Infof("Bind <%s,%s> created.", svr.Addr, c.Name)
}

// Select return a server using spec loadbalance
func (c *Cluster) Select(req *fasthttp.Request) string {
	c.rwLock.RLock()
	defer c.rwLock.RUnlock()

	index := c.lb.Select(req, c.svrs)

	if 0 > index {
		return ""
	}

	e := util.Get(c.svrs, index)

	if nil == e {
		return ""
	}

	s, _ := e.Value.(string)

	return s
}

// Matches return true if req matches
func (c *Cluster) Matches(req *fasthttp.Request) bool {
	return c.regexp.MatchString(string(req.URI().Path()))
}

// Marshal marshal
func (c *Cluster) Marshal() []byte {
	v, _ := json.Marshal(c)
	return v
}

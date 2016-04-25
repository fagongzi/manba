package model

import (
	"container/list"
	"encoding/json"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/pkg/lb"
	"github.com/fagongzi/gateway/pkg/util"
	"io"
	"net/http"
	"regexp"
	"sync"
)

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

func UnMarshalCluster(data []byte) *Cluster {
	v := &Cluster{}
	err := json.Unmarshal(data, v)

	if err != nil {
		return v
	}

	c, _ := NewCluster(v.Name, v.Pattern, v.LbName)

	return c
}

func UnMarshalClusterFromReader(r io.Reader) (*Cluster, error) {
	v := &Cluster{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	if nil != err {
		return nil, err
	}

	return v, v.init()
}

func NewCluster(name string, pattern string, lbName string) (*Cluster, error) {
	c := &Cluster{
		Name:    name,
		Pattern: pattern,
		LbName:  lbName,
	}

	return c, c.init()
}

func (self *Cluster) init() error {
	reg, err := regexp.Compile(self.Pattern)

	if nil != err {
		return err
	}

	self.regexp = reg
	self.svrs = list.New()
	self.lb = lb.NewLoadBalance(self.LbName)
	self.rwLock = &sync.RWMutex{}

	return nil
}

func (self *Cluster) updateFrom(cluster *Cluster) {
	if self.rwLock != nil {
		self.rwLock.Lock()
		defer self.rwLock.Unlock()
	}

	self.Pattern = cluster.Pattern
	self.LbName = cluster.LbName

	self.regexp, _ = regexp.Compile(self.Pattern)
	self.lb = lb.NewLoadBalance(self.LbName)

	log.Infof("Cluster <%s> updated, %+v", self.Name, self)
}

func (self *Cluster) doInEveryBindServers(callback func(string)) {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	for iter := self.svrs.Back(); iter != nil; iter = iter.Prev() {
		addr, _ := iter.Value.(string)
		callback(addr)
	}
}

func (self *Cluster) unbind(svr *Server) {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	self.doUnBind(svr)
}

func (self *Cluster) doUnBind(svr *Server) {
	util.Remove(self.svrs, svr.Addr)
	log.Infof("UnBind <%s,%s> succ.", svr.Addr, self.Name)
}

func (self *Cluster) bind(svr *Server) {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	if util.IndexOf(self.svrs, svr.Addr) >= 0 {
		log.Infof("Bind <%s,%s> already created.", svr.Addr, self.Name)
		return
	}

	self.svrs.PushBack(svr.Addr)

	log.Infof("Bind <%s,%s> created.", svr.Addr, self.Name)
}

func (self *Cluster) Select(req *http.Request) string {
	self.rwLock.RLock()
	defer self.rwLock.RUnlock()

	index := self.lb.Select(req, self.svrs)

	if 0 > index {
		return ""
	}

	e := util.Get(self.svrs, index)

	if nil == e {
		return ""
	}

	s, _ := e.Value.(string)

	return s
}

func (self *Cluster) Matches(req *http.Request) bool {
	return self.regexp.MatchString(req.URL.Path)
}

func (self *Cluster) Marshal() []byte {
	v, _ := json.Marshal(self)
	return v
}

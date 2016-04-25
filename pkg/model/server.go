package model

import (
	"encoding/json"
	"fmt"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	CHECK_SUCCESS = "OK"
)

type Status int
type Circuit int

const (
	DOWN = Status(0)
	UP   = Status(1)
)

const (
	CIRCUIT_OPEN  = Circuit(0)
	CIRCUIT_HALF  = Circuit(1)
	CIRCUIT_CLOSE = Circuit(2)
)

const (
	DEFAULT_CHECK_INTERVAL = 5
	DEFAULT_CHECK_TIMEOUT  = 3

	THRESHOLD_MAX_FAIED_COUNT = 10
)

type Server struct {
	Schema string `json:"schema,omitempty"`
	Addr   string `json:"addr,omitempty"` // ip:port

	CheckPath     string `json:"checkPath,omitempty"`     // begin with / checkpath, expect return OK.
	CheckDuration int    `json:"checkDuration,omitempty"` // check interval, unit second
	CheckTimeout  int    `json:"checkTimeout,omitempty"`
	Status        Status `json:"status,omitempty"` // Server status

	MaxQPS          int `json:"maxQPS,omitempty"`
	HalfToOpen      int `json:"halfToOpen,omitempty"`
	HalfTrafficRate int `json:"halfTrafficRate,omitempty"`
	CloseCount      int `json:"closeCount,omitempty"`

	BindClusters []string `json:"bindClusters,omitempty"`

	httpClient       *http.Client
	checkFailCount   int
	prevStatus       Status
	useCheckDuration int

	circuit Circuit
	lock    *sync.Mutex

	checkStopped bool
}

func UnMarshalServer(data []byte) *Server {
	v := &Server{}
	json.Unmarshal(data, v)

	if 0 == v.CheckTimeout {
		v.CheckTimeout = DEFAULT_CHECK_TIMEOUT
	}

	return v
}

func UnMarshalServerFromReader(r io.Reader) (*Server, error) {
	v := &Server{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	v.Status = DOWN

	if 0 == v.CheckTimeout {
		v.CheckTimeout = DEFAULT_CHECK_TIMEOUT
	}

	return v, err
}

func (self *Server) updateFrom(svr *Server) {
	if self.lock != nil {
		self.Lock()
		defer self.UnLock()
	}

	self.MaxQPS = svr.MaxQPS
	self.HalfToOpen = svr.HalfToOpen
	self.HalfTrafficRate = svr.HalfTrafficRate
	self.CloseCount = svr.CloseCount

	log.Infof("Server <%s> updated, %+v", self.Addr, self)
}

func (self *Server) Marshal() []byte {
	v, _ := json.Marshal(self)
	return v
}

func (self *Server) GetCircuit() Circuit {
	return self.circuit
}

func (self *Server) OpenCircuit() {
	self.circuit = CIRCUIT_OPEN
}

func (self *Server) CloseCircuit() {
	self.circuit = CIRCUIT_CLOSE
}

func (self *Server) HalfCircuit() {
	self.circuit = CIRCUIT_HALF
}

func (self *Server) Lock() {
	self.lock.Lock()
}

func (self *Server) UnLock() {
	self.lock.Unlock()
}

func (self *Server) init() {
	self.httpClient = &http.Client{
		Timeout: time.Second * self.getCheckTimeout(),
	}

	self.circuit = CIRCUIT_OPEN
	self.lock = &sync.Mutex{}
	self.checkStopped = false
}

func (self *Server) stopCheck() {
	self.checkStopped = true
}

func (self *Server) getCheckTimeout() time.Duration {
	if self.CheckTimeout == 0 {
		return time.Duration(DEFAULT_CHECK_TIMEOUT)
	} else {
		return time.Duration(self.CheckTimeout)
	}
}

func (self *Server) check(cb func(*Server)) bool {
	succ := false
	defer func() {
		if succ {
			self.reset()
		} else {
			self.fail()
		}

		if !self.checkStopped {
			cb(self)
		}
	}()

	log.Debugf("Server <%s, %s> start check.", self.Addr, self.CheckPath)

	resp, err := self.httpClient.Get(self.getCheckURL())

	if err != nil {
		log.Warnf("Server <%s, %s, %d> check fail.", self.Addr, self.CheckPath, self.checkFailCount+1)
		return succ
	}

	defer resp.Body.Close()

	if http.StatusOK != resp.StatusCode {
		log.Warnf("Server <%s, %s, %d, %d> check fail.", self.Addr, self.CheckPath, resp.StatusCode, self.checkFailCount+1)
		return succ
	}

	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return false
	}

	succ = string(body) == CHECK_SUCCESS
	return succ
}

func (self *Server) getCheckURL() string {
	return fmt.Sprintf("%s://%s%s", self.Schema, self.Addr, self.CheckPath)
}

func (self *Server) fail() {
	self.checkFailCount += 1
	self.useCheckDuration += self.useCheckDuration / 2
}

func (self *Server) reset() {
	self.checkFailCount = 0
	self.useCheckDuration = self.CheckDuration
}

func (self *Server) changeTo(status Status) {
	self.prevStatus = self.Status
	self.Status = status
}

func (self *Server) statusChanged() bool {
	return self.prevStatus != self.Status
}

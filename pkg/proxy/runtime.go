package proxy

import (
	"container/list"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fagongzi/gateway/pkg/lb"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/goetty"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/collection"
	"github.com/valyala/fasthttp"
)

type clusterRuntime struct {
	meta *metapb.Cluster
	svrs *list.List
	lb   lb.LoadBalance
}

func newClusterRuntime(meta *metapb.Cluster) *clusterRuntime {
	rt := &clusterRuntime{
		meta: meta,
		svrs: list.New(),
		lb:   lb.NewLoadBalance(meta.LoadBalance),
	}

	return rt
}

func (c *clusterRuntime) updateMeta(meta *metapb.Cluster) {
	c.meta = meta
	c.lb = lb.NewLoadBalance(meta.LoadBalance)
}

func (c *clusterRuntime) foreach(do func(uint64)) {
	for iter := c.svrs.Back(); iter != nil; iter = iter.Prev() {
		addr, _ := iter.Value.(uint64)
		do(addr)
	}
}

func (c *clusterRuntime) remove(id uint64) {
	collection.Remove(c.svrs, id)
	log.Infof("bind <%d,%d> removed.", c.meta.ID, id)
}

func (c *clusterRuntime) add(id uint64) {
	if collection.IndexOf(c.svrs, id) >= 0 {
		log.Warnf("bind <%d,%d> already created.", c.meta.ID, id)
		return
	}

	c.svrs.PushBack(id)
	log.Infof("bind <%d,%d> created.", c.meta.ID, id)
}

func (c *clusterRuntime) selectServer(req *fasthttp.Request) uint64 {
	index := c.lb.Select(req, c.svrs)
	if 0 > index {
		return 0
	}

	e := collection.Get(c.svrs, index)
	if nil == e {
		return 0
	}

	id, _ := e.Value.(uint64)
	return id
}

type serverRuntime struct {
	tw               *goetty.TimeoutWheel
	meta             *metapb.Server
	status           metapb.Status
	heathTimeout     goetty.Timeout
	checkFailCount   int
	useCheckDuration time.Duration
	circuit          metapb.CircuitStatus
}

func newServerRuntime(meta *metapb.Server, tw *goetty.TimeoutWheel) *serverRuntime {
	rt := &serverRuntime{
		meta:    meta,
		tw:      tw,
		status:  metapb.Down,
		circuit: metapb.Open,
	}

	return rt
}

func (s *serverRuntime) updateMeta(meta *metapb.Server) {
	s.meta = meta
}

func (s *serverRuntime) getCheckURL() string {
	return fmt.Sprintf("%s://%s%s", strings.ToLower(s.meta.Protocol.String()), s.meta.Addr, s.meta.HeathCheck.Path)
}

func (s *serverRuntime) fail() {
	s.checkFailCount++
	s.useCheckDuration += s.useCheckDuration / 2
}

func (s *serverRuntime) reset() {
	s.checkFailCount = 0
	s.useCheckDuration = time.Duration(s.meta.HeathCheck.CheckInterval)
}

func (s *serverRuntime) changeTo(status metapb.Status) {
	s.status = status
}

func (s *serverRuntime) isCircuitStatus(target metapb.CircuitStatus) bool {
	return s.circuit == target
}

func (s *serverRuntime) circuitToClose() {
	if s.meta.CircuitBreaker == nil ||
		s.circuit == metapb.Close {
		return
	}

	s.circuit = metapb.Close
	log.Warnf("server <%s> change to close", s.meta.ID)
	s.tw.Schedule(time.Duration(s.meta.CircuitBreaker.CloseTimeout), s.circuitToHalf, nil)
}

func (s *serverRuntime) circuitToOpen() {
	if s.meta.CircuitBreaker == nil ||
		s.circuit == metapb.Open ||
		s.circuit != metapb.Half {
		return
	}

	s.circuit = metapb.Open
	log.Infof("server <%s> change to open", s.meta.ID)
}

func (s *serverRuntime) circuitToHalf(arg interface{}) {
	if s.meta.CircuitBreaker != nil {
		s.circuit = metapb.Open
		log.Warnf("server <%s> change to half", s.meta.ID)
	}
}

type ipSegment struct {
	value []string
}

func parseFrom(value string) *ipSegment {
	ip := &ipSegment{}
	ip.value = strings.Split(value, ".")
	return ip
}

func (ip *ipSegment) matches(value string) bool {
	tmp := strings.Split(value, ".")

	for index, v := range ip.value {
		if v != "*" && v != tmp[index] {
			return false
		}
	}

	return true
}

type apiValidation struct {
	meta  *metapb.Validation
	rules []*apiRule
}

type apiRule struct {
	pattern *regexp.Regexp
}

type apiNode struct {
	meta        *metapb.DispatchNode
	validations []*apiValidation
}

type apiRuntime struct {
	meta            *metapb.API
	nodes           []*apiNode
	urlPattern      *regexp.Regexp
	defaultCookies  []*fasthttp.Cookie
	parsedWhitelist []*ipSegment
	parsedBlacklist []*ipSegment
}

func newAPIRuntime(meta *metapb.API) *apiRuntime {
	ar := &apiRuntime{
		meta: meta,
	}
	ar.init()

	return ar
}

func (a *apiRuntime) updateMeta(meta *metapb.API) {
	a.meta = meta
	a.init()
}

func (a *apiRuntime) init() {
	if a.meta.URLPattern != "" {
		a.urlPattern = regexp.MustCompile(a.meta.URLPattern)
	}

	a.nodes = make([]*apiNode, 0)
	for _, n := range a.meta.Nodes {
		rn := &apiNode{
			meta: n,
		}
		a.nodes = append(a.nodes, rn)

		for _, v := range n.Validations {
			rv := &apiValidation{
				meta: v,
			}

			for _, r := range v.Rules {
				rv.rules = append(rv.rules, &apiRule{
					pattern: regexp.MustCompile(r.Expression),
				})
			}

			rn.validations = append(rn.validations, rv)
		}
	}

	a.defaultCookies = make([]*fasthttp.Cookie, 0)
	if nil != a.meta.DefaultValue {
		for _, c := range a.meta.DefaultValue.Cookies {
			ck := &fasthttp.Cookie{}
			ck.SetKey(c.Name)
			ck.SetValue(c.Value)
			a.defaultCookies = append(a.defaultCookies, ck)
		}
	}

	a.parsedWhitelist = make([]*ipSegment, 0)
	a.parsedBlacklist = make([]*ipSegment, 0)
	if nil != a.meta.IPAccessControl {
		if a.meta.IPAccessControl.Whitelist != nil {
			for _, ip := range a.meta.IPAccessControl.Whitelist {
				a.parsedWhitelist = append(a.parsedWhitelist, parseFrom(ip))
			}
		}

		if a.meta.IPAccessControl.Blacklist != nil {
			for _, ip := range a.meta.IPAccessControl.Blacklist {
				a.parsedBlacklist = append(a.parsedBlacklist, parseFrom(ip))
			}
		}
	}

	return
}

func (a *apiRuntime) allowWithBlacklist(ip string) bool {
	if a.meta.IPAccessControl == nil {
		return true
	}

	for _, i := range a.parsedBlacklist {
		if i.matches(ip) {
			return false
		}
	}

	return true
}

func (a *apiRuntime) allowWithWhitelist(ip string) bool {
	if a.meta.IPAccessControl == nil || len(a.meta.IPAccessControl.Whitelist) == 0 {
		return true
	}

	for _, i := range a.parsedWhitelist {
		if i.matches(ip) {
			return true
		}
	}

	return false
}

func (a *apiRuntime) renderDefault(ctx *fasthttp.RequestCtx) {
	if a.meta.DefaultValue == nil {
		return
	}

	for _, header := range a.meta.DefaultValue.Headers {
		ctx.Response.Header.Add(header.Name, header.Value)
	}

	for _, ck := range a.defaultCookies {
		ctx.Response.Header.SetCookie(ck)
	}

	ctx.Write(a.meta.DefaultValue.Body)
}

func (a *apiRuntime) rewriteURL(req *fasthttp.Request, rewrite string) string {
	if rewrite == "" || a.meta.URLPattern == "" {
		return ""
	}

	return a.urlPattern.ReplaceAllString(string(req.URI().RequestURI()), rewrite)
}

func (a *apiRuntime) matches(req *fasthttp.Request) bool {
	return a.isUp() &&
		(a.isDomainMatches(req) ||
			(a.isMethodMatches(req) && a.isURIMatches(req)))
}

func (a *apiRuntime) isUp() bool {
	return a.meta.Status == metapb.Up
}

func (a *apiRuntime) isMethodMatches(req *fasthttp.Request) bool {
	return a.meta.Method == "*" || strings.ToUpper(string(req.Header.Method())) == a.meta.Method
}

func (a *apiRuntime) isURIMatches(req *fasthttp.Request) bool {
	return a.urlPattern.Match(req.URI().RequestURI())
}

func (a *apiRuntime) isDomainMatches(req *fasthttp.Request) bool {
	return a.meta.Domain != "" && string(req.Header.Host()) == a.meta.Domain
}

func (a *apiNode) validate(req *fasthttp.Request) bool {
	if len(a.validations) == 0 {
		return true
	}

	for _, v := range a.validations {
		if !v.validate(req) {
			return false
		}
	}

	return true
}

func (v *apiValidation) validate(req *fasthttp.Request) bool {
	if len(v.rules) == 0 && !v.meta.Required {
		return true
	}

	value := paramValue(&v.meta.Parameter, req)
	if "" == value && v.meta.Required {
		return false
	} else if "" == value && !v.meta.Required {
		return true
	}

	for _, r := range v.rules {
		if !r.validate([]byte(value)) {
			return false
		}
	}

	return true
}

func (r *apiRule) validate(value []byte) bool {
	return r.pattern.Match(value)
}

type routingRuntime struct {
	meta *metapb.Routing
	rand *rand.Rand
}

func newRoutingRuntime(meta *metapb.Routing) *routingRuntime {
	r := &routingRuntime{
		meta: meta,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	r.init()

	return r
}

func (a *routingRuntime) updateMeta(meta *metapb.Routing) {
	a.meta = meta
	a.init()
}

func (a *routingRuntime) init() {
	return
}

func (a *routingRuntime) matches(req *fasthttp.Request) bool {
	for _, c := range a.meta.Conditions {
		if !conditionsMatches(&c, req) {
			return false
		}
	}

	n := a.rand.Intn(100)
	return n < int(a.meta.TrafficRate)
}

func conditionsMatches(cond *metapb.Condition, req *fasthttp.Request) bool {
	attrValue := paramValue(&cond.Parameter, req)
	if attrValue == "" {
		return false
	}

	switch cond.Cmp {
	case metapb.CMPEQ:
		return eq(attrValue, cond.Expect)
	case metapb.CMPLT:
		return lt(attrValue, cond.Expect)
	case metapb.CMPLE:
		return le(attrValue, cond.Expect)
	case metapb.CMPGT:
		return gt(attrValue, cond.Expect)
	case metapb.CMPGE:
		return ge(attrValue, cond.Expect)
	case metapb.CMPIn:
		return in(attrValue, cond.Expect)
	case metapb.CMPMatch:
		return reg(attrValue, cond.Expect)
	default:
		return false
	}
}

func eq(attrValue string, expect string) bool {
	return attrValue == expect
}

func lt(attrValue string, expect string) bool {
	s, err := strconv.Atoi(attrValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(expect)
	if err != nil {
		return false
	}

	return s < t
}

func le(attrValue string, expect string) bool {
	s, err := strconv.Atoi(attrValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(expect)
	if err != nil {
		return false
	}

	return s <= t
}

func gt(attrValue string, expect string) bool {
	s, err := strconv.Atoi(attrValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(expect)
	if err != nil {
		return false
	}

	return s > t
}

func ge(attrValue string, expect string) bool {
	s, err := strconv.Atoi(attrValue)
	if err != nil {
		return false
	}

	t, err := strconv.Atoi(expect)
	if err != nil {
		return false
	}

	return s >= t
}

func in(attrValue string, expect string) bool {
	return strings.Index(expect, attrValue) != -1
}

func reg(attrValue string, expect string) bool {
	matches, err := regexp.MatchString(expect, attrValue)
	return err == nil && matches
}

func paramValue(param *metapb.Parameter, req *fasthttp.Request) string {
	switch param.Source {
	case metapb.QueryString:
		return getQueryValue(param.Name, req)
	case metapb.FormData:
		return getFormValue(param.Name, req)
	case metapb.JSONBody:
		return getJSONBodyValue(param.Name, req)
	case metapb.Header:
		return getHeaderValue(param.Name, req)
	case metapb.Cookie:
		return getCookieValue(param.Name, req)
	default:
		return ""
	}
}

func getCookieValue(name string, req *fasthttp.Request) string {
	return string(req.Header.Cookie(name))
}

func getHeaderValue(name string, req *fasthttp.Request) string {
	return string(req.Header.Peek(name))
}

func getQueryValue(name string, req *fasthttp.Request) string {
	v, _ := url.QueryUnescape(string(req.URI().QueryArgs().Peek(name)))
	return v
}

func getFormValue(name string, req *fasthttp.Request) string {
	return string(req.PostArgs().Peek(name))
}

func getJSONBodyValue(name string, req *fasthttp.Request) string {
	data := make(map[string]interface{})
	err := json.Unmarshal(req.Body(), &data)
	if err != nil {
		return ""
	}

	var value interface{}
	for _, name := range strings.Split(name, ".") {
		if value == nil {
			value = data[name]
			continue
		}

		if m, ok := value.(map[string]interface{}); ok {
			value = m[name]
		} else {
			return ""
		}
	}

	if ret, ok := value.(string); ok {
		return ret
	}

	ret, _ := json.Marshal(&value)
	return string(ret)
}

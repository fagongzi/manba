package proxy

import (
	"container/list"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	"github.com/fagongzi/gateway/pkg/lb"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/goetty"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/collection"
	"github.com/fagongzi/util/hack"
	pbutil "github.com/fagongzi/util/protoc"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
)

var (
	dependP = regexp.MustCompile(`\$\w+\.\w+`)
)

type clusterRuntime struct {
	meta *metapb.Cluster
	svrs *list.List
	lb   lb.LoadBalance
}

func newClusterRuntime(meta *metapb.Cluster) *clusterRuntime {
	return &clusterRuntime{
		meta: meta,
		svrs: list.New(),
		lb:   lb.NewLoadBalance(meta.LoadBalance),
	}
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
	log.Infof("bind <%d,%d> inactived", c.meta.ID, id)
}

func (c *clusterRuntime) add(id uint64) {
	if collection.IndexOf(c.svrs, id) >= 0 {
		return
	}

	c.svrs.PushBack(id)
	log.Infof("bind <%d,%d> actived", c.meta.ID, id)
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
	sync.RWMutex

	tw               *goetty.TimeoutWheel
	limiter          *rate.Limiter
	meta             *metapb.Server
	status           metapb.Status
	heathTimeout     goetty.Timeout
	checkFailCount   int
	useCheckDuration time.Duration
	circuit          metapb.CircuitStatus
}

func newServerRuntime(meta *metapb.Server, tw *goetty.TimeoutWheel) *serverRuntime {
	rt := &serverRuntime{
		tw:      tw,
		status:  metapb.Down,
		circuit: metapb.Open,
	}

	rt.updateMeta(meta)
	return rt
}

func (s *serverRuntime) clone() *serverRuntime {
	meta := &metapb.Server{}
	pbutil.MustUnmarshal(meta, pbutil.MustMarshal(s.meta))
	return newServerRuntime(meta, s.tw)
}

func (s *serverRuntime) updateMeta(meta *metapb.Server) {
	s.heathTimeout.Stop()
	tw := s.tw
	*s = serverRuntime{}
	s.tw = tw
	s.meta = meta
	s.limiter = rate.NewLimiter(rate.Every(time.Second/time.Duration(meta.MaxQPS)), int(meta.MaxQPS))
	s.status = metapb.Down
	s.circuit = metapb.Open
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

func (s *serverRuntime) getCircuitStatus() metapb.CircuitStatus {
	s.RLock()
	value := s.circuit
	s.RUnlock()
	return value
}

func (s *serverRuntime) circuitToClose() {
	s.Lock()
	if s.meta.CircuitBreaker == nil ||
		s.circuit == metapb.Close {
		s.Unlock()
		return
	}

	s.circuit = metapb.Close
	log.Warnf("server <%d> change to close", s.meta.ID)
	s.tw.Schedule(time.Duration(s.meta.CircuitBreaker.CloseTimeout), s.circuitToHalf, nil)
	s.Unlock()
}

func (s *serverRuntime) circuitToOpen() {
	s.Lock()
	if s.meta.CircuitBreaker == nil ||
		s.circuit == metapb.Open ||
		s.circuit != metapb.Half {
		s.Unlock()
		return
	}

	s.circuit = metapb.Open
	log.Infof("server <%d> change to open", s.meta.ID)
	s.Unlock()
}

func (s *serverRuntime) circuitToHalf(arg interface{}) {
	s.Lock()
	if s.meta.CircuitBreaker != nil {
		s.circuit = metapb.Half
		log.Warnf("server <%d> change to half", s.meta.ID)
	}
	s.Unlock()
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
	meta              *metapb.DispatchNode
	validations       []*apiValidation
	defaultCookies    []*fasthttp.Cookie
	dependencies      []string
	dependenciesPaths [][]string
}

func newAPINode(meta *metapb.DispatchNode) *apiNode {
	rn := &apiNode{
		meta: meta,
	}

	if meta.URLRewrite != "" {
		matches := dependP.FindAllStringSubmatch(meta.URLRewrite, -1)
		for _, match := range matches {
			rn.dependencies = append(rn.dependencies, match[0])
			rn.dependenciesPaths = append(rn.dependenciesPaths, strings.Split(match[0][1:], "."))
		}
	}

	if nil != meta.DefaultValue {
		for _, c := range meta.DefaultValue.Cookies {
			ck := &fasthttp.Cookie{}
			ck.SetKey(c.Name)
			ck.SetValue(c.Value)
			rn.defaultCookies = append(rn.defaultCookies, ck)
		}
	}

	for _, v := range meta.Validations {
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

	return rn
}

func (n *apiNode) clone() *apiNode {
	meta := &metapb.DispatchNode{}
	pbutil.MustUnmarshal(meta, pbutil.MustMarshal(n.meta))
	return newAPINode(meta)
}

func (n *apiNode) validate(req *fasthttp.Request) bool {
	if len(n.validations) == 0 {
		return true
	}

	for _, v := range n.validations {
		if !v.validate(req) {
			return false
		}
	}

	return true
}

type renderAttr struct {
	meta     *metapb.RenderAttr
	extracts [][]string
}

type renderObject struct {
	meta  *metapb.RenderObject
	attrs []*renderAttr
}

type apiRuntime struct {
	meta                *metapb.API
	nodes               []*apiNode
	urlPattern          *regexp.Regexp
	defaultCookies      []*fasthttp.Cookie
	parsedWhitelist     []*ipSegment
	parsedBlacklist     []*ipSegment
	parsedRenderObjects []*renderObject
}

func newAPIRuntime(meta *metapb.API) *apiRuntime {
	ar := &apiRuntime{
		meta: meta,
	}
	ar.init()

	return ar
}

func (a *apiRuntime) clone() *apiRuntime {
	meta := &metapb.API{}
	pbutil.MustUnmarshal(meta, pbutil.MustMarshal(a.meta))
	return newAPIRuntime(meta)
}

func (a *apiRuntime) updateMeta(meta *metapb.API) {
	*a = apiRuntime{}
	a.meta = meta
	a.init()
}

func (a *apiRuntime) compare(i, j int) bool {
	return a.nodes[i].meta.BatchIndex-a.nodes[j].meta.BatchIndex < 0
}

func (a *apiRuntime) init() {
	if a.meta.URLPattern != "" {
		a.urlPattern = regexp.MustCompile(a.meta.URLPattern)
	}

	for _, n := range a.meta.Nodes {
		a.nodes = append(a.nodes, newAPINode(n))
	}

	sort.Slice(a.nodes, a.compare)

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

	if nil != a.meta.RenderTemplate {
		for _, obj := range a.meta.RenderTemplate.Objects {
			rob := &renderObject{
				meta: obj,
			}

			for _, attr := range obj.Attrs {
				rattr := &renderAttr{
					meta: attr,
				}
				rob.attrs = append(rob.attrs, rattr)

				extracts := strings.Split(attr.ExtractExp, ",")
				for _, extract := range extracts {
					rattr.extracts = append(rattr.extracts, strings.Split(extract, "."))
				}
			}

			a.parsedRenderObjects = append(a.parsedRenderObjects, rob)
		}
	}

	return
}

func (a *apiRuntime) hasRenderTemplate() bool {
	return a.meta.RenderTemplate != nil
}

func (a *apiRuntime) hasDefaultValue() bool {
	return a.meta.DefaultValue != nil
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

func (a *apiRuntime) rewriteURL(req *fasthttp.Request, node *apiNode, ctx *multiContext) string {
	rewrite := node.meta.URLRewrite
	if rewrite == "" || a.meta.URLPattern == "" {
		return ""
	}

	if nil != ctx && len(node.dependencies) > 0 {
		for idx, dep := range node.dependencies {
			rewrite = strings.Replace(rewrite, dep, ctx.getAttr(node.dependenciesPaths[idx]...), -1)
		}
	}

	return a.urlPattern.ReplaceAllString(hack.SliceToString(req.URI().RequestURI()), rewrite)
}

func (a *apiRuntime) matches(req *fasthttp.Request) bool {
	if !a.isUp() {
		return false
	}

	switch a.matchRule() {
	case metapb.MatchAll:
		return a.isDomainMatches(req) && a.isMethodMatches(req) && a.isURIMatches(req)
	case metapb.MatchAny:
		return a.isDomainMatches(req) || a.isMethodMatches(req) || a.isURIMatches(req)
	default:
		return a.isDomainMatches(req) || (a.isMethodMatches(req) && a.isURIMatches(req))
	}
}

func (a *apiRuntime) isUp() bool {
	return a.meta.Status == metapb.Up
}

func (a *apiRuntime) isMethodMatches(req *fasthttp.Request) bool {
	return a.meta.Method == "*" || strings.ToUpper(hack.SliceToString(req.Header.Method())) == a.meta.Method
}

func (a *apiRuntime) isURIMatches(req *fasthttp.Request) bool {
	if a.urlPattern == nil {
		return false
	}

	return a.urlPattern.Match(req.URI().RequestURI())
}

func (a *apiRuntime) isDomainMatches(req *fasthttp.Request) bool {
	return a.meta.Domain != "" && hack.SliceToString(req.Header.Host()) == a.meta.Domain
}

func (a *apiRuntime) position() uint32 {
	return a.meta.GetPosition()
}

func (a *apiRuntime) matchRule() metapb.MatchRule {
	return a.meta.GetMatchRule()
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
		if !r.validate(hack.StringToSlice(value)) {
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
	return &routingRuntime{
		meta: meta,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (a *routingRuntime) updateMeta(meta *metapb.Routing) {
	a.meta = meta
}

func (a *routingRuntime) matches(apiID uint64, req *fasthttp.Request) bool {
	if a.meta.API > 0 && apiID != a.meta.API {
		return false
	}

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
		value, _, _, err := jsonparser.Get(req.Body(), param.Name)
		if err != nil {
			return ""
		}
		return hack.SliceToString(value)
	case metapb.Header:
		return getHeaderValue(param.Name, req)
	case metapb.Cookie:
		return getCookieValue(param.Name, req)
	case metapb.PathValue:
		return getPathValue(int(param.Index), req)
	default:
		return ""
	}
}

func getCookieValue(name string, req *fasthttp.Request) string {
	return hack.SliceToString(req.Header.Cookie(name))
}

func getHeaderValue(name string, req *fasthttp.Request) string {
	return hack.SliceToString(req.Header.Peek(name))
}

func getQueryValue(name string, req *fasthttp.Request) string {
	v, _ := url.QueryUnescape(hack.SliceToString(req.URI().QueryArgs().Peek(name)))
	return v
}

func getPathValue(idx int, req *fasthttp.Request) string {
	path := hack.SliceToString(req.URI().Path()[1:])
	values := strings.Split(path, "/")
	if len(values) <= idx {
		return ""
	}

	return values[idx]
}

func getFormValue(name string, req *fasthttp.Request) string {
	return string(req.PostArgs().Peek(name))
}

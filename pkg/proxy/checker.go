package proxy

import (
	"context"
	"time"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/log"
	"github.com/valyala/fasthttp"
)

func (r *dispatcher) readyToHeathChecker() {
	for i := 0; i < r.cnf.Option.LimitCountHeathCheckWorker; i++ {
		r.runner.RunCancelableTask(func(ctx context.Context) {
			log.Infof("start server check worker")

			for {
				select {
				case <-ctx.Done():
					return
				case id := <-r.checkerC:
					r.check(id)
				}
			}
		})
	}
}

func (r *dispatcher) addToCheck(svr *serverRuntime) {
	svr.circuit = metapb.Open
	svr.status = metapb.Unknown
	if svr.meta.HeathCheck != nil {
		svr.useCheckDuration = time.Duration(svr.meta.HeathCheck.CheckInterval)
	}
	svr.heathTimeout.Stop()
	r.checkerC <- svr.meta.ID
}

func (r *dispatcher) heathCheckTimeout(arg interface{}) {
	r.checkerC <- arg.(uint64)
}

func (r *dispatcher) check(id uint64) {
	r.Lock()
	defer r.Unlock()

	svr, ok := r.servers[id]
	if !ok {
		return
	}

	defer func() {
		if svr.meta.HeathCheck != nil {
			if svr.useCheckDuration > r.cnf.Option.LimitIntervalHeathCheck {
				svr.useCheckDuration = r.cnf.Option.LimitIntervalHeathCheck
			}
			svr.heathTimeout, _ = r.tw.Schedule(svr.useCheckDuration, r.heathCheckTimeout, id)
		}
	}()

	prev := svr.status

	if svr.meta.HeathCheck == nil {
		log.Warnf("server <%d> heath check not setting", svr.meta.ID)
		svr.changeTo(metapb.Up)
	} else {

		if r.doCheck(svr) {
			svr.changeTo(metapb.Up)
		} else {
			svr.changeTo(metapb.Down)
		}
	}

	if prev != svr.status {
		clusters, ok := r.binds[svr.meta.ID]

		if svr.status == metapb.Up {
			log.Infof("server <%d> UP",
				svr.meta.ID)

			if ok {
				for _, c := range clusters {
					c.add(svr.meta.ID)
				}
			}
		} else {
			log.Infof("server <%d> DOWN",
				svr.meta.ID)

			if ok {
				for _, c := range clusters {
					c.remove(svr.meta.ID)
				}
			}
		}
	}
}

func (r *dispatcher) doCheck(svr *serverRuntime) bool {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(svr.getCheckURL())

	opt := util.DefaultHTTPOption()
	opt.ReadTimeout = time.Duration(svr.meta.HeathCheck.Timeout)
	resp, err := r.httpClient.Do(req, svr.meta.Addr, opt)
	defer fasthttp.ReleaseResponse(resp)
	if err != nil {
		log.Warnf("server <%d, %s, %d> check failed, errors:\n%+v",
			svr.meta.ID,
			svr.meta.HeathCheck.Path,
			svr.checkFailCount+1,
			err)
		svr.fail()
		return false
	}

	if fasthttp.StatusOK != resp.StatusCode() {
		log.Warnf("server <%d, %s, %d, %d> check failed",
			svr.meta.ID,
			svr.meta.HeathCheck.Path,
			resp.StatusCode(),
			svr.checkFailCount+1)
		svr.fail()
		return false
	}

	if svr.meta.HeathCheck.Body != "" &&
		svr.meta.HeathCheck.Body != string(resp.Body()) {
		log.Warnf("server <%s, %s, %d> check failed, body <%s>, expect <%s>",
			svr.meta.Addr,
			svr.meta.HeathCheck.Path,
			svr.checkFailCount+1,
			resp.Body(),
			svr.meta.HeathCheck.Body)
		svr.fail()
		return false
	}

	svr.reset()
	return true
}

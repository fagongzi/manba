package proxy

import (
	"context"
	"time"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/store"
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
	if svr.meta.HeathCheck != nil {
		svr.useCheckDuration = time.Duration(svr.meta.HeathCheck.CheckInterval)
	}
	svr.heathTimeout.Stop()
	r.checkerC <- svr.meta.ID
}

func (r *dispatcher) heathCheckTimeout(arg interface{}) {
	id := arg.(uint64)
	if _, ok := r.servers[id]; ok {
		r.checkerC <- id
	}
}

func (r *dispatcher) check(id uint64) {
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

	status := metapb.Unknown
	prev := r.getServerStatus(svr.meta.ID)

	if svr.meta.HeathCheck == nil {
		log.Warnf("server <%d> heath check not setting", svr.meta.ID)
		r.watchEventC <- &store.Evt{
			Src:  eventSrcStatusChanged,
			Type: eventTypeStatusChanged,
			Value: statusChanged{
				meta:   *svr.meta,
				status: metapb.Up,
			},
		}
	} else {
		if r.doCheck(svr) {
			status = metapb.Up
		} else {
			status = metapb.Down
		}
	}

	if prev != status {
		r.watchEventC <- &store.Evt{
			Src:  eventSrcStatusChanged,
			Type: eventTypeStatusChanged,
			Value: statusChanged{
				meta:   *svr.meta,
				status: status,
			},
		}
	}
}

func (r *dispatcher) doCheck(svr *serverRuntime) bool {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(svr.getCheckURL())

	opt := util.DefaultHTTPOption()
	*opt = *globalHTTPOptions
	opt.ReadTimeout = time.Duration(svr.meta.HeathCheck.Timeout)

	resp, err := r.httpClient.Do(req, svr.meta.Addr, opt)
	defer fasthttp.ReleaseResponse(resp)
	if err != nil {
		log.Warnf("server <%d, %s, %d> check failed, errors:\n%+v",
			svr.meta.ID,
			svr.getCheckURL(),
			svr.checkFailCount+1,
			err)
		svr.fail()
		return false
	}

	if fasthttp.StatusOK != resp.StatusCode() {
		log.Warnf("server <%d, %s, %d, %d> check failed",
			svr.meta.ID,
			svr.getCheckURL(),
			resp.StatusCode(),
			svr.checkFailCount+1)
		svr.fail()
		return false
	}

	if svr.meta.HeathCheck.Body != "" &&
		svr.meta.HeathCheck.Body != string(resp.Body()) {
		log.Warnf("server <%s, %s, %d> check failed, body <%s>, expect <%s>",
			svr.meta.Addr,
			svr.getCheckURL(),
			svr.checkFailCount+1,
			resp.Body(),
			svr.meta.HeathCheck.Body)
		svr.fail()
		return false
	}

	svr.reset()
	return true
}

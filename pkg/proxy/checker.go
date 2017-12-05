package proxy

import (
	"context"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/log"
	"github.com/valyala/fasthttp"
)

func (r *dispatcher) readyToHeathChecker() {
	for i := 0; i < r.cnf.HeathCheckWorkerCount; i++ {
		r.taskRunner.RunCancelableTask(func(ctx context.Context) {
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
	svr.circuit = model.CircuitOpen
	svr.prevStatus = model.Down
	svr.status = model.Down
	svr.useCheckDuration = svr.meta.HeathCheck.Interval

	r.checkerC <- svr.meta.ID
}

func (r *dispatcher) heathCheckTimeout(arg interface{}) {
	r.checkerC <- arg.(string)
}

func (r *dispatcher) check(id string) {
	r.RLock()
	defer r.RUnlock()

	svr, ok := r.servers[id]
	if !ok {
		return
	}

	defer func() {
		if svr.useCheckDuration > r.cnf.MaxServerCheckDuration {
			svr.useCheckDuration = r.cnf.MaxServerCheckDuration
		}
		r.tw.Schedule(svr.useCheckDuration, r.heathCheckTimeout, id)
	}()

	if svr.meta.External {
		log.Warnf("server <%s> heath check is using external", svr.meta.ID)
		svr.changeTo(model.Up)
	} else if svr.meta.HeathCheck == nil {
		log.Warnf("server <%s> heath check not setting", svr.meta.ID)
		svr.changeTo(model.Up)
	} else {
		if r.doCheck(svr) {
			svr.changeTo(model.Up)
			if svr.statusChanged() {
				log.Infof("meta: server <%s> UP",
					svr.meta.ID)
			}
		} else {
			svr.changeTo(model.Down)
			if svr.statusChanged() {
				log.Warnf("meta: server <%s, %s> DOWN",
					svr.meta.ID,
					svr.meta.HeathCheck.Path)
			}
		}
	}

	if svr.statusChanged() {
		binded := r.mapping[svr.meta.ID]

		if svr.status == model.Up {
			log.Infof("meta: server <%s> UP",
				svr.meta.ID)

			for _, c := range binded {
				c.add(svr.meta.ID)
			}
		} else {
			log.Infof("meta: server <%s> DOWN",
				svr.meta.ID)

			for _, c := range binded {
				c.remove(svr.meta.ID)
			}
		}
	}
}

func (r *dispatcher) doCheck(svr *serverRuntime) bool {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(svr.getCheckURL())

	opt := util.DefaultHTTPOption()
	opt.ReadTimeout = svr.meta.HeathCheck.Timeout
	resp, err := r.httpClient.Do(req, svr.meta.Addr, opt)
	defer fasthttp.ReleaseResponse(resp)
	if err != nil {
		log.Warnf("server <%s, %s, %d> check failed, errors:\n%+v",
			svr.meta.ID,
			svr.meta.HeathCheck.Path,
			svr.checkFailCount+1,
			err)
		svr.fail()
		return false
	}

	if fasthttp.StatusOK != resp.StatusCode() {
		log.Warnf("server <%s, %s, %d, %d> check failed",
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

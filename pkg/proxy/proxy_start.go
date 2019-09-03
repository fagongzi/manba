package proxy

import (
	"net"
	"net/http"
	"sync/atomic"

	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/log"
	"github.com/soheilhy/cmux"
	"github.com/valyala/fasthttp"
)

// Start start proxy
func (p *Proxy) Start() {
	go p.listenToStop()

	p.startMetrics()
	p.startReadyTasks()

	if !p.cfg.Option.EnableWebSocket {
		go p.startHTTPS()
		p.startHTTP()

		return
	}

	go p.startHTTPSCMUX()
	p.startHTTPCMUX()
}

// Stop stop the proxy
func (p *Proxy) Stop() {
	log.Infof("stop: start to stop gateway proxy")

	p.stopWG.Add(1)
	p.stopC <- struct{}{}
	p.stopWG.Wait()

	log.Infof("stop: gateway proxy stopped")
}

func (p *Proxy) listenToStop() {
	<-p.stopC
	p.doStop()
}

func (p *Proxy) doStop() {
	p.stopOnce.Do(func() {
		defer p.stopWG.Done()
		p.setStopped()
		p.runner.Stop()
	})
}

func (p *Proxy) stopRPC() error {
	return p.rpcListener.Close()
}

func (p *Proxy) setStopped() {
	atomic.StoreInt32(&p.stopped, 1)
}

func (p *Proxy) isStopped() bool {
	return atomic.LoadInt32(&p.stopped) == 1
}

func (p *Proxy) startMetrics() {
	util.StartMetricsPush(p.runner, p.cfg.Metric)
}

func (p *Proxy) startReadyTasks() {
	p.readyToGCJSEngine()
	p.readyToCopy()
	p.readyToDispatch()
}

func (p *Proxy) newHTTPServer() *fasthttp.Server {
	return &fasthttp.Server{
		Handler:                       p.ServeFastHTTP,
		ReadBufferSize:                p.cfg.Option.LimitBufferRead,
		WriteBufferSize:               p.cfg.Option.LimitBufferWrite,
		MaxRequestBodySize:            p.cfg.Option.LimitBytesBody,
		DisableHeaderNamesNormalizing: p.cfg.Option.DisableHeaderNameNormalizing,
	}
}

func (p *Proxy) startHTTP() {
	log.Infof("start http at %s", p.cfg.Addr)
	s := p.newHTTPServer()
	err := s.ListenAndServe(p.cfg.Addr)
	if err != nil {
		log.Fatalf("start http listeners failed with %+v", err)
	}
}

func (p *Proxy) startHTTPWithListener(l net.Listener) {
	log.Infof("start http at %s", p.cfg.Addr)
	s := p.newHTTPServer()
	err := s.Serve(l)
	if err != nil {
		log.Fatalf("start http listeners failed with %+v", err)
	}
}

func (p *Proxy) startHTTPS() {
	if !p.enableHTTPS() {
		return
	}

	defaultCertData, defaultKeyData := p.mustParseDefaultTLSCert()

	log.Infof("start https at %s", p.cfg.AddrHTTPS)
	s := p.newHTTPServer()
	p.appendCertsEmbed(s, defaultCertData, defaultKeyData)
	err := s.ListenAndServeTLS(p.cfg.AddrHTTPS, "", "")
	if err != nil {
		log.Fatalf("start http listeners failed with %+v", err)
	}
}

func (p *Proxy) startHTTPSWithListener(l net.Listener) {
	defaultCertData, defaultKeyData := p.mustParseDefaultTLSCert()

	log.Infof("start https at %s", p.cfg.AddrHTTPS)
	s := p.newHTTPServer()
	p.appendCertsEmbed(s, defaultCertData, defaultKeyData)
	err := s.ServeTLS(l, "", "")
	if err != nil {
		log.Fatalf("start http listeners failed with %+v", err)
	}
}

func (p *Proxy) startHTTPWebSocketWithListener(l net.Listener) {
	log.Infof("start http websocket at %s", p.cfg.Addr)
	s := &http.Server{
		Handler: p,
	}
	err := s.Serve(l)
	if err != nil {
		log.Fatalf("start http websocket failed with %+v", err)
	}
}

func (p *Proxy) startHTTPSWebSocketWithListener(l net.Listener) {
	defaultCertData, defaultKeyData := p.mustParseDefaultTLSCert()

	log.Infof("start https websocket at %s", p.cfg.Addr)
	s := &http.Server{
		Handler: p,
	}
	p.configTLSConfig(s, defaultCertData, defaultKeyData)
	err := s.ServeTLS(l, "", "")
	if err != nil {
		log.Fatalf("start https websocket failed with errors %+v", err)
	}
}

func (p *Proxy) startHTTPCMUX() {
	l, err := net.Listen("tcp", p.cfg.Addr)
	if err != nil {
		log.Fatalf("start http failed failed with %+v",
			err)
	}

	m := cmux.New(l)
	go p.startHTTPWithListener(m.Match(cmux.Any()))
	go p.startHTTPWebSocketWithListener(m.Match(cmux.HTTP1HeaderField("Upgrade", "websocket")))
	err = m.Serve()
	if err != nil {
		log.Fatalf("start http failed failed with %+v",
			err)
	}
}

func (p *Proxy) startHTTPSCMUX() {
	if !p.enableHTTPS() {
		return
	}

	l, err := net.Listen("tcp", p.cfg.AddrHTTPS)
	if err != nil {
		log.Fatalf("start https failed failed with %+v",
			err)
	}

	m := cmux.New(l)
	go p.startHTTPSWithListener(m.Match(cmux.Any()))
	go p.startHTTPSWebSocketWithListener(m.Match(cmux.HTTP1HeaderField("Upgrade", "websocket")))
	err = m.Serve()
	if err != nil {
		log.Fatalf("start https failed failed with %+v",
			err)
	}
}

package model

import (
	"fmt"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/toolkits/net"
	"strings"
	"time"
)

var (
	TICKER = time.Second * 3
	TTL    = uint64(5)
)

func (self EtcdStore) Registry(proxyInfo *ProxyInfo) error {
	timer := time.NewTicker(TICKER)

	go func() {
		for {
			<-timer.C
			log.Debug("Registry start")
			self.doRegistry(proxyInfo)
		}
	}()

	return nil
}

func (self EtcdStore) doRegistry(proxyInfo *ProxyInfo) {
	proxyInfo.Conf.Addr = convertIp(proxyInfo.Conf.Addr)
	proxyInfo.Conf.MgrAddr = convertIp(proxyInfo.Conf.MgrAddr)

	key := fmt.Sprintf("%s/%s", self.proxiesDir, proxyInfo.Conf.Addr)
	_, err := self.cli.Set(key, proxyInfo.Marshal(), TTL)

	if err != nil {
		log.ErrorError(err, "Registry fail.")
	}
}

func (self EtcdStore) GetProxies() ([]*ProxyInfo, error) {
	rsp, err := self.cli.Get(self.proxiesDir, true, false)

	if nil != err {
		return nil, err
	}

	l := rsp.Node.Nodes.Len()
	proxies := make([]*ProxyInfo, l)

	for i := 0; i < l; i++ {
		proxies[i] = UnMarshalProxyInfo([]byte(rsp.Node.Nodes[i].Value))
	}

	return proxies, nil
}

func (self EtcdStore) ChangeLogLevel(addr string, level string) error {
	rpcClient, _ := net.RpcClient("tcp", addr, time.Second*5)

	req := SetLogReq{
		Level: level,
	}

	rsp := &SetLogRsp{
		Code: 0,
	}

	return rpcClient.Call("Manager.SetLogLevel", req, rsp)
}

func (self EtcdStore) AddAnalysisPoint(proxyAddr, serverAddr string, secs int) error {
	rpcClient, _ := net.RpcClient("tcp", proxyAddr, time.Second*5)

	req := AddAnalysisPointReq{
		Addr: serverAddr,
		Secs: secs,
	}

	rsp := &AddAnalysisPointRsp{
		Code: 0,
	}

	return rpcClient.Call("Manager.AddAnalysisPoint", req, rsp)
}

func (self EtcdStore) GetAnalysisPoint(proxyAddr, serverAddr string, secs int) (*GetAnalysisPointRsp, error) {
	rpcClient, err := net.RpcClient("tcp", proxyAddr, time.Second*5)

	if nil != err {
		return nil, err
	}

	req := GetAnalysisPointReq{
		Addr: serverAddr,
		Secs: secs,
	}

	rsp := &GetAnalysisPointRsp{}

	err = rpcClient.Call("Manager.GetAnalysisPoint", req, rsp)

	return rsp, err
}

func convertIp(addr string) string {
	if strings.HasPrefix(addr, ":") {
		ips, err := net.IntranetIP()

		if err == nil {
			addr = strings.Replace(addr, ":", fmt.Sprintf("%s:", ips[0]), 1)
		}
	}

	return addr
}

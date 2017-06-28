package model

import (
	"fmt"
	"time"

	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/toolkits/net"
)

// Registry registry self
func (e EtcdStore) Registry(proxyInfo *ProxyInfo) error {
	timer := time.NewTicker(TICKER)

	go func() {
		for {
			<-timer.C
			log.Debug("Registry start")
			e.doRegistry(proxyInfo)
		}
	}()

	return nil
}

func (e EtcdStore) doRegistry(proxyInfo *ProxyInfo) {
	proxyInfo.Conf.Addr = convertIP(proxyInfo.Conf.Addr)
	proxyInfo.Conf.MgrAddr = convertIP(proxyInfo.Conf.MgrAddr)

	key := fmt.Sprintf("%s/%s", e.proxiesDir, proxyInfo.Conf.Addr)
	_, err := e.cli.Set(key, proxyInfo.Marshal(), TTL)

	if err != nil {
		log.ErrorError(err, "Registry fail.")
	}
}

// GetProxies return runable proxies
func (e EtcdStore) GetProxies() ([]*ProxyInfo, error) {
	rsp, err := e.cli.Get(e.proxiesDir, true, false)

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

// ChangeLogLevel change proxy log level
func (e EtcdStore) ChangeLogLevel(addr string, level string) error {
	rpcClient, _ := net.RpcClient("tcp", addr, time.Second*5)

	req := SetLogReq{
		Level: level,
	}

	rsp := &SetLogRsp{
		Code: 0,
	}

	return rpcClient.Call("Manager.SetLogLevel", req, rsp)
}

// AddAnalysisPoint add a analysis point
func (e EtcdStore) AddAnalysisPoint(proxyAddr, serverAddr string, secs int) error {
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

// GetAnalysisPoint return analysis point data
func (e EtcdStore) GetAnalysisPoint(proxyAddr, serverAddr string, secs int) (*GetAnalysisPointRsp, error) {
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

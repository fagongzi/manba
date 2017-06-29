package model

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/fagongzi/log"
	"github.com/toolkits/net"
)

// Registry registry self
func (e *EtcdStore) Registry(proxyInfo *ProxyInfo) error {
	timer := time.NewTicker(TICKER)

	e.taskRunner.RunCancelableTask(func(ctx context.Context) {
		select {
		case <-ctx.Done():
			timer.Stop()
		case <-timer.C:
			e.doRegistry(proxyInfo)
		}
	})

	return nil
}

func (e *EtcdStore) doRegistry(proxyInfo *ProxyInfo) {
	proxyInfo.Conf.Addr = convertIP(proxyInfo.Conf.Addr)
	proxyInfo.Conf.MgrAddr = convertIP(proxyInfo.Conf.MgrAddr)

	key := fmt.Sprintf("%s/%s", e.proxiesDir, proxyInfo.Conf.Addr)
	err := e.putTTL(key, string(proxyInfo.Marshal()), TTL)

	if err != nil {
		log.Errorf("store: registry failed, errors:\n%+v",
			err)
	}
}

// GetProxies return runable proxies
func (e *EtcdStore) GetProxies() ([]*ProxyInfo, error) {
	var values []*ProxyInfo
	err := e.getList(e.proxiesDir, func(item *mvccpb.KeyValue) {
		values = append(values, UnMarshalProxyInfo(item.Value))
	})

	return values, err
}

// ChangeLogLevel change proxy log level
func (e *EtcdStore) ChangeLogLevel(addr string, level string) error {
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
func (e *EtcdStore) AddAnalysisPoint(proxyAddr, serverAddr string, secs int) error {
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
func (e *EtcdStore) GetAnalysisPoint(proxyAddr, serverAddr string, secs int) (*GetAnalysisPointRsp, error) {
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

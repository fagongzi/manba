package model

import (
	"context"
	"fmt"
	"time"

	"github.com/fagongzi/log"
	"github.com/hashicorp/consul/api"
	"github.com/toolkits/net"
)

// Registry registry self
func (s *consulStore) Registry(proxyInfo *ProxyInfo) error {
	timer := time.NewTicker(TICKER)

	s.taskRunner.RunCancelableTask(func(ctx context.Context) {
		select {
		case <-ctx.Done():
			timer.Stop()
		case <-timer.C:
			s.doRegistry(proxyInfo)
		}
	})

	return nil
}

func (s *consulStore) doRegistry(proxyInfo *ProxyInfo) {
	proxyInfo.Conf.Addr = convertIP(proxyInfo.Conf.Addr)
	proxyInfo.Conf.MgrAddr = convertIP(proxyInfo.Conf.MgrAddr)

	key := fmt.Sprintf("%s/%s", s.proxiesDir, proxyInfo.Conf.Addr)
	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: []byte(proxyInfo.Marshal()),
	}, nil)

	if err != nil {
		log.Errorf("store: registry failed, errors:\n%+v",
			err)
	}
}

// GetProxies return runable proxies
func (s *consulStore) GetProxies() ([]*ProxyInfo, error) {
	pairs, _, err := s.client.KV().List(s.proxiesDir, nil)

	if nil != err {
		return nil, err
	}

	values := make([]*ProxyInfo, len(pairs))
	i := 0

	for _, pair := range pairs {
		values[i] = UnMarshalProxyInfo(pair.Value)
		i++
	}

	return values, nil
}

// ChangeLogLevel change proxy log level
func (s *consulStore) ChangeLogLevel(addr string, level string) error {
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
func (s *consulStore) AddAnalysisPoint(proxyAddr, serverAddr string, secs int) error {
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
func (s *consulStore) GetAnalysisPoint(proxyAddr, serverAddr string, secs int) (*GetAnalysisPointRsp, error) {
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

package main

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

func createRouting() error {
	c, err := getClient()
	if err != nil {
		return err
	}

	rb := c.NewRoutingBuilder()
	// 必选项
	// 下线
	rb.Down()
	// 上线
	rb.Up()
	// 拆分10%的流量到cluster 2,剩余90%的流量按照原有规则转发
	rb.TrafficRate(10)
	rb.To(2)
	rb.Strategy(metapb.Split)

	// 可选项
	// 目标请求必须包含 v 的query string，且必须是v1，那么就是把v1的流量导流10%到cluster 2
	param := metapb.Parameter{
		Name:   "v",
		Source: metapb.QueryString,
	}
	rb.AddCondition(param, metapb.CMPEQ, "v1")

	_, err = rb.Commit()
	return err
}

func updateRouting(id uint64) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	routing, err := c.GetRouting(id)
	if err != nil {
		return err
	}

	rb := c.NewRoutingBuilder().Use(*routing)
	// 必选项
	// 下线
	rb.Down()
	// 上线
	rb.Up()
	// 拆分10%的流量到cluster 2,剩余90%的流量按照原有规则转发
	rb.TrafficRate(10)
	rb.To(2)
	rb.Strategy(metapb.Split)

	// 可选项
	// 目标请求必须包含 v 的query string，且必须是v1，那么就是把v1的流量导流10%到cluster 2
	param := metapb.Parameter{
		Name:   "v",
		Source: metapb.QueryString,
	}
	rb.AddCondition(param, metapb.CMPEQ, "v1")

	_, err = rb.Commit()
	return err
}

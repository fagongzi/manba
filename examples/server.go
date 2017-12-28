package main

import (
	"fmt"
	"time"
)

func createServer() error {
	c, err := getClient()
	if err != nil {
		return err
	}

	sb := c.NewServerBuilder()
	// 必选项
	sb.Addr("127.0.0.1:8080").HTTPBackend().MaxQPS(100)

	// 健康检查，可选项
	// 每个10秒钟检查一次，每次检查的超时时间30秒，即30秒后端Server没有返回认为后端不健康
	sb.CheckHTTPCode("/check/path", time.Second*10, time.Second*30)

	// 熔断器，可选项
	// 统计周期1秒钟
	sb.CircuitBreakerCheckPeriod(time.Second)
	// 在Close状态60秒后自动转到Half状态
	sb.CircuitBreakerCloseToHalfTimeout(time.Second * 60)
	// Half状态下，允许10%的流量流入后端
	sb.CircuitBreakerHalfTrafficRate(10)
	// 在Half状态，1秒内有2%的请求失败了，转换到Close状态
	sb.CircuitBreakerHalfToCloseCondition(2)
	// 在Half状态，1秒内有90%的请求成功了，转换到Open状态
	sb.CircuitBreakerHalfToOpenCondition(90)

	id, err := sb.Commit()
	if err != nil {
		return err
	}

	fmt.Printf("server id is: %d", id)

	// 把这个server加入到cluster 1
	c.AddBind(1, id)

	// 把这个server从cluster 1 移除
	c.RemoveBind(1, id)

	// 加入到cluster 2
	c.AddBind(2, id)
	return nil
}

func updateServer(id uint64) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	svr, err := c.GetServer(id)
	if err != nil {
		return err
	}

	sb := c.NewServerBuilder()
	sb.Use(*svr)

	// 修改你想要修改的字段
	sb.MaxQPS(1000)
	sb.NoCircuitBreaker() // 删除熔断器
	sb.NoHeathCheck()     // 删除健康检查

	_, err = sb.Commit()
	return err
}

func deleteServer(id uint64) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	return c.RemoveServer(id)
}

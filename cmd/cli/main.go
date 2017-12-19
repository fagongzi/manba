package main

import (
	"fmt"
	"time"

	"github.com/fagongzi/gateway/pkg/client"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

func main() {
	cli, err := client.NewClientWithEtcdDiscovery("/services", time.Second*30, "http://127.0.0.1:2379")
	if err != nil {
		fmt.Println(err)
		return
	}

	api, _ := cli.GetAPI(1001)
	api.Nodes[0].URLRewrite = "/$1"

	id, err := cli.PutAPI(*api)
	if err != nil {
		panic(err)
	}
	fmt.Printf("api: %d\n", id)
}

func addAPI(cli client.Client) {
	api := client.BuildAPI("test", metapb.Up,
		client.WithDefaultResult([]byte("default value"), nil, nil),
		client.WithMatchMethod("GET"),
		client.WithMatchURL("/api/(.*)"),
		client.WithAddDispatchNode(1,
			client.WithURLWriteDispatch("$1")))

	id, err := cli.PutAPI(api)
	if err != nil {
		panic(err)
	}
	fmt.Printf("api: %d\n", id)
}

func removeBind(cluster, server uint64, cli client.Client) {
	err := cli.RemoveBind(cluster, server)
	if err != nil {
		panic(err)
	}
	fmt.Printf("remove bind: %d,%d\n", cluster, server)
}

func addBind(cluster, server uint64, cli client.Client) {
	err := cli.AddBind(cluster, server)
	if err != nil {
		panic(err)
	}
	fmt.Printf("add bind: %d,%d\n", cluster, server)
}

func addCluster(cli client.Client) {
	id, err := cli.PutCluster(metapb.Cluster{
		Name:        "c1",
		LoadBalance: metapb.RoundRobin,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("cluster: %d\n", id)
}

func addServer(cli client.Client) {
	svr := client.BuildServer("127.0.0.1:9090",
		client.WithCheckHTTPBody("/check", "OK", time.Second*10, time.Second*10),
		client.WithHTTPBackend(),
		client.WithQPS(1))
	id, err := cli.PutServer(svr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("server: %d\n", id)
}

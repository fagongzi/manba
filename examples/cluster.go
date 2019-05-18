package main

import (
	"fmt"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

func createCluster() error {
	c, err := getClient()
	if err != nil {
		return err
	}

	id, err := c.NewClusterBuilder().Name("cluster-01").Loadbalance(metapb.RoundRobin).Commit()
	if err != nil {
		return err
	}

	fmt.Printf("cluster id is: %d", id)
	return nil
}

func deleteCluster(id uint64) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	return c.RemoveCluster(id)
}

func updateCluster(id uint64) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	cluster, err := c.GetCluster(id)
	if err != nil {
		return err
	}

	// 修改名称
	_, err = c.NewClusterBuilder().Use(*cluster).Name("cluster-1").Commit()
	if err != nil {
		return err
	}

	fmt.Printf("cluster %d name is updated", id)
	return nil
}

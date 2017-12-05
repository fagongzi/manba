package client

import (
	"fmt"

	"github.com/fagongzi/gateway/pkg/model"
)

const (
	logTag = "gateway-http-client"

	httpGET    = "GET"
	httpPUT    = "PUT"
	httpPOST   = "POST"
	httpDELETE = "DELETE"

	apiCluster = "clusters"
	apiServer  = "servers"
)

func apiForResources(version APIVersion, resource string) string {
	return fmt.Sprintf("api/%s/%s", version, resource)
}

func apiForResource(version APIVersion, resource string, id string) string {
	return fmt.Sprintf("%s/%s", apiForResources(version, resource), id)
}

// Client is a gateway client that used manager the meta data of gateway
type Client interface {
	// AddCluster add a cluster
	AddCluster(cluster *model.Cluster) (string, error)
	// DeleteCluster delete a cluster
	DeleteCluster(id string) error
	// GetCluster returns a cluster
	GetCluster(id string) (*model.Cluster, error)
	// GetClusters returns a all clusters
	GetClusters() ([]*model.Cluster, error)

	// AddServer add a server
	AddServer(server *model.Server) (string, error)
}

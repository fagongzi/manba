package model

import (
	"encoding/json"
	"fmt"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/pkg/plugin"
)

const (
	// ActionClusters query cluster
	ActionClusters = "/clusters"
	// ActionServers query servers
	ActionServers = "/clusters/"
)

// Clusters clusters
type Clusters struct {
	Clusters map[string]*Cluster `json:"clusters"`
}

// Servers servers
type Servers struct {
	Servers map[string]*Server `json:"servers"`
}

// ServiceDiscoveryDriver service discovery driver
// service discovery is used for auto discovery domains, servers and binds.
type ServiceDiscoveryDriver struct {
	registryCenter *plugin.RegistryCenter
}

// NewServiceDiscoveryDriver new ServiceDiscoveryDriver
func NewServiceDiscoveryDriver(registryCenter *plugin.RegistryCenter) *ServiceDiscoveryDriver {
	return &ServiceDiscoveryDriver{
		registryCenter: registryCenter,
	}
}

// GetAllClusters get all clusters
func (d *ServiceDiscoveryDriver) GetAllClusters() (*Clusters, error) {
	data, err := d.registryCenter.DoGet(plugin.TypeServiceDiscovery, ActionClusters)
	if err != nil {
		return nil, err
	}

	log.Infof("Plugin<service-discovery> get clusters, resp is<%s>", string(data))

	clusters := &Clusters{}
	err = json.Unmarshal(data, clusters)
	if err != nil {
		return nil, err
	}

	return clusters, nil
}

// GetServersByClusterName get all servers in domains
func (d *ServiceDiscoveryDriver) GetServersByClusterName(clusterName string) (*Servers, error) {
	data, err := d.registryCenter.DoGet(plugin.TypeServiceDiscovery, fmt.Sprintf("%s%s", ActionServers, clusterName))
	if err != nil {
		return nil, err
	}

	log.Infof("Plugin<service-discovery> get servers by cluster<%s>, resp is<%s>", clusterName, string(data))

	servers := &Servers{}
	err = json.Unmarshal(data, servers)
	if err != nil {
		return nil, err
	}

	return servers, nil
}

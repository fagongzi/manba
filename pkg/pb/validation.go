package pb

import (
	"fmt"

	"github.com/fagongzi/gateway/pkg/plugin"

	"github.com/fagongzi/gateway/pkg/expr"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

// ValidateRouting validate routing
func ValidateRouting(value *metapb.Routing) error {
	if value.API == 0 {
		return fmt.Errorf("missing api")
	}

	if value.ClusterID == 0 {
		return fmt.Errorf("missing cluster")
	}

	if value.Name == "" {
		return fmt.Errorf("missing name")
	}

	if value.TrafficRate <= 0 || value.TrafficRate > 100 {
		return fmt.Errorf("error traffic rate: %d", value.TrafficRate)
	}

	return nil
}

// ValidateCluster validate cluster
func ValidateCluster(value *metapb.Cluster) error {
	if value.Name == "" {
		return fmt.Errorf("missing name")
	}

	return nil
}

// ValidateServer validate server
func ValidateServer(value *metapb.Server) error {
	if value.Addr == "" {
		return fmt.Errorf("missing server address")
	}

	if value.MaxQPS == 0 {
		return fmt.Errorf("missing server max qps")
	}

	return nil
}

// ValidateAPI validate api
func ValidateAPI(value *metapb.API) error {
	if value.Name == "" {
		return fmt.Errorf("missing api name")
	}

	if value.URLPattern == "" {
		return fmt.Errorf("missing URLPattern")
	}

	for _, n := range value.Nodes {
		if n.URLRewrite != "" {
			_, err := expr.Parse([]byte(n.URLRewrite))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ValidatePlugin validate plugin
func ValidatePlugin(value *metapb.Plugin) error {
	if value.Name == "" {
		return fmt.Errorf("missing plugin name")
	}

	if value.Version == 0 {
		return fmt.Errorf("missing plugin version")
	}

	if len(value.Content) == 0 {
		return fmt.Errorf("missing plugin content")
	}

	_, err := plugin.NewRuntime(string(value.Content), string(value.Content))
	if err != nil {
		return err
	}

	return nil
}

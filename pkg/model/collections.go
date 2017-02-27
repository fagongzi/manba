package model

// MinusClusters get minus collection
// return collection which element is in c1 and not in c, c1 - c
func MinusClusters(c map[string]*Cluster, c1 map[string]*Cluster, filter func(*Cluster) bool) map[string]*Cluster {
	value := make(map[string]*Cluster)

	for _, e1 := range c1 {
		if !filter(e1) {
			continue
		}

		if _, ok := c[e1.Name]; !ok {
			value[e1.Name] = e1
		}
	}

	return value
}

// MinusServers get minus collection
// return collection which element is in c1 and not in c
func MinusServers(c map[string]*Server, c1 map[string]*Server, filter func(*Server) bool) map[string]*Server {
	value := make(map[string]*Server)

	for _, e1 := range c1 {
		if !filter(e1) {
			continue
		}

		if _, ok := c[e1.Addr]; !ok {
			value[e1.Addr] = e1
		}
	}

	return value
}

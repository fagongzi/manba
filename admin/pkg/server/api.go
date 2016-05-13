package server

func (self *AdminServer) initApiRoute() {
	self.e.Get("/api/lbs", self.getLbs())

	self.e.Get("/api/proxies", self.getProxies())
	self.e.Post("/api/proxies/:addr/:level", self.changeLogLevel())

	self.e.Get("/api/clusters", self.getClusters())
	self.e.Get("/api/clusters/:id", self.getCluster())
	self.e.Delete("/api/clusters/:id", self.deleteCluster())
	self.e.Post("/api/clusters", self.newCluster())
	self.e.Put("/api/clusters", self.updateCluster())

	self.e.Get("/api/servers", self.getServers())
	self.e.Get("/api/servers/:id", self.getServer())
	self.e.Delete("/api/servers/:id", self.deleteServer())
	self.e.Post("/api/servers", self.newServer())
	self.e.Put("/api/servers", self.updateServer())

	self.e.Post("/api/binds", self.newBind())
	self.e.Delete("/api/binds", self.unBind())

	self.e.Get("/api/aggregations", self.getAggregations())
	self.e.Post("/api/aggregations", self.newAggregation())
	self.e.Delete("/api/aggregations", self.deleteAggregation())

	self.e.Get("/api/routings", self.getRoutings())
	self.e.Post("/api/routings", self.newRouting())

	self.e.Get("/api/analysis/:proxy/:server/:secs", self.getAnalysis())
	self.e.Post("/api/analysis", self.newAnalysis())
}

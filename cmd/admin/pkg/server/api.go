package server

func (server *AdminServer) initAPIRoute() {
	server.e.Get("/api/lbs", server.getLbs())

	server.e.Get("/api/proxies", server.getProxies())
	server.e.Post("/api/proxies/:addr/:level", server.changeLogLevel())

	server.e.Get("/api/clusters", server.getClusters())
	server.e.Get("/api/clusters/:id", server.getCluster())
	server.e.Delete("/api/clusters/:id", server.deleteCluster())
	server.e.Post("/api/clusters", server.newCluster())
	server.e.Put("/api/clusters", server.updateCluster())

	server.e.Get("/api/servers", server.getServers())
	server.e.Get("/api/servers/:id", server.getServer())
	server.e.Delete("/api/servers/:id", server.deleteServer())
	server.e.Post("/api/servers", server.newServer())
	server.e.Put("/api/servers", server.updateServer())

	server.e.Post("/api/binds", server.newBind())
	server.e.Delete("/api/binds", server.unBind())

	server.e.Get("/api/aggregations", server.getAggregations())
	server.e.Post("/api/aggregations", server.newAggregation())
	server.e.Delete("/api/aggregations", server.deleteAggregation())

	server.e.Get("/api/routings", server.getRoutings())
	server.e.Post("/api/routings", server.newRouting())

	server.e.Get("/api/analysis/:proxy/:server/:secs", server.getAnalysis())
	server.e.Post("/api/analysis", server.newAnalysis())
}

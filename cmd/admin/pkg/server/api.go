package server

func (server *AdminServer) initAPIRoute() {
	server.e.GET("/api/lbs", server.getLbs())

	server.e.GET("/api/proxies", server.getProxies())
	server.e.POST("/api/proxies/:addr/:level", server.changeLogLevel())

	server.e.GET("/api/clusters", server.getClusters())
	server.e.GET("/api/clusters/:id", server.getCluster())
	server.e.DELETE("/api/clusters/:id", server.deleteCluster())
	server.e.POST("/api/clusters", server.newCluster())
	server.e.PUT("/api/clusters", server.updateCluster())

	server.e.GET("/api/servers", server.getServers())
	server.e.GET("/api/servers/:id", server.getServer())
	server.e.DELETE("/api/servers/:id", server.deleteServer())
	server.e.POST("/api/servers", server.newServer())
	server.e.PUT("/api/servers", server.updateServer())

	server.e.POST("/api/binds", server.newBind())
	server.e.DELETE("/api/binds", server.unBind())

	server.e.GET("/api/aggregations", server.getAggregations())
	server.e.GET("/api/aggregation", server.getAggregation())
	server.e.POST("/api/aggregations", server.newAggregation())
	server.e.PUT("/api/aggregations", server.updateAggregation())
	server.e.DELETE("/api/aggregations", server.deleteAggregation())

	server.e.GET("/api/routings", server.getRoutings())
	server.e.POST("/api/routings", server.newRouting())

	server.e.GET("/api/analysis/:proxy/:server/:secs", server.getAnalysis())
	server.e.POST("/api/analysis", server.newAnalysis())
}

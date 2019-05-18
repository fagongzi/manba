package grpcx

// API api
type API struct {
	Name string
	HTTP APIEntrypoint
}

// APIEntrypoint api http entrypoint
type APIEntrypoint struct {
	GET, PUT, DELETE, POST string
}

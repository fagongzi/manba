package service

import (
	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
	"github.com/fagongzi/gateway/pkg/store"
)

var (
	// MetaService global service
	MetaService rpcpb.MetaServiceServer
	// Store global store db
	Store store.Store
)

// Init init service package
func Init(db store.Store) {
	Store = db
	MetaService = newMetaService(db)
}

ApiServer
--------------
ApiServer provides GRPC APIs to manage metadata of Gateway which is Cluster,Server, Routing, and API info.

# APIs Exposed
```
// MetaService is a interface for meta manager
service MetaService {
    rpc PutCluster        (PutClusterReq)        returns (PutClusterRsp)         {}
    rpc RemoveCluster     (RemoveClusterReq)     returns (RemoveClusterRsp)      {}
    rpc GetCluster        (GetClusterReq)        returns (GetClusterRsp)         {}
    rpc GetClusterList    (GetClusterListReq)    returns (stream metapb.Cluster) {}
      
    rpc PutServer         (PutServerReq)         returns (PutServerRsp)          {}
    rpc RemoveServer      (RemoveServerReq)      returns (RemoveServerRsp)       {}
    rpc GetServer         (GetServerReq)         returns (GetServerRsp)          {}
    rpc GetServerList     (GetServerListReq)     returns (stream metapb.Server)  {}
      
    rpc PutAPI            (PutAPIReq)            returns (PutAPIRsp)             {}
    rpc RemoveAPI         (RemoveAPIReq)         returns (RemoveAPIRsp)          {}
    rpc GetAPI            (GetAPIReq)            returns (GetAPIRsp)             {}
    rpc GetAPIList        (GetAPIListReq)        returns (stream metapb.API)     {}
      
    rpc PutRouting        (PutRoutingReq)        returns (PutRoutingRsp)         {}
    rpc RemoveRouting     (RemoveRoutingReq)     returns (RemoveRoutingRsp)      {}
    rpc GetRouting        (GetRoutingReq)        returns (GetRoutingRsp)         {}
    rpc GetRoutingList    (GetRoutingListReq)    returns (stream metapb.Routing) {}
      
    rpc AddBind           (AddBindReq)           returns (AddBindRsp)            {}
    rpc RemoveBind        (RemoveBindReq)        returns (RemoveBindRsp)         {}
    rpc RemoveClusterBind (RemoveClusterBindReq) returns (RemoveClusterBindRsp)  {}
    rpc GetBindServers    (GetBindServersReq)    returns (GetBindServersRsp)     {}
}
```
The PB is under `pkg/pb/rpcpb`.

# Client
For the moment, Gateway supports Go client. Here are [examples](../examples) of Go clients of Gateway which manage metadata.

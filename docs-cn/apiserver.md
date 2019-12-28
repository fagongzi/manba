ApiServer
--------------
ApiServer对外提供GRPC接口，用来管理Manba的元信息（Cluster、Server、Routing以及API）。

# 对外开放的接口：
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
具体的PB在项目的`pkg/pb/rpcpb`目录下

# 客户端
目前Gateway支持GO的客户端，这里以Gateway的GO客户端管理元信息的例子，参见[examples](../examples)

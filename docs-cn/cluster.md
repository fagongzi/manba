Cluster
-------
在Gateway中，Cluster是一个逻辑的概念。它是后端真实Server的一个逻辑组，在同一个组内的后端Server提供相同的服务。

# 增删改查
## 创建Cluster
一个Cluster包含2部分信息:

* Cluster Name
  Cluster名称，全局唯一.

* Cluster Load Balance
  Cluster采取的负载均衡算法.

一旦在`Admin`中创建了一个Cluster，所有正在运行的Proxy会通过Ectd的watch机制感知到，并且实时生效，你可以在Proxy中看到对应的日志。

## 更新cluster
目前只能更新cluster的负载均衡算法，一旦一个Cluster被更新，所有的Proxy通过Ectd的watch机制感知到，并且实时生效。

## 删除cluster
一旦一个Cluster被删除，所有的Proxy通过Ectd的watch机制感知到，并且实时生效。**注意，如果Clsuter上绑定了Server，那么Cluster是不能被删除的，需要先解绑Server**。

# 绑定后端Server
一个Cluster可以绑定多个Server,Cluster和Server的绑定关系发生变化的时候，所有的Proxy通过Ectd的watch机制感知到，并且实时生效。
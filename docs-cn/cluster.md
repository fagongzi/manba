Cluster
-------
在Gateway中，Cluster是一个逻辑的概念。它是后端真实Server的一个逻辑组，在同一个组内的后端Server提供相同的服务。

# Cluster属性
一个Cluster包含2部分信息:
## ID
Cluster ID, 全局唯一。

## Name
Cluster名称

## LoadBalance
Cluster采取的负载均衡算法。
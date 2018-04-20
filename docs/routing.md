Routing
------
在Gateway中，Routing代表一个路由，利用路由我们可以实现我们的AB Test以及线上导流等高级特性。

# Routing属性
## ID
Routing ID，唯一标识。

## clusterID
流量路由到哪一个Cluster。

## apiID
针对哪一个API

## Condition（可选）
路由条件，当满足这些条件，则Gateway执行这个路由。路由条件可以设置`cookie`、`querystring`、`header`、`json body`,`path value`中的参数的表达式。不配置，匹配所有流量。

## RoutingStrategy
路由策略，目前支持`Split`分发。分发是指：把满足条件的请求按照比例转发到目标Cluster，剩余比例的流量按照正常流程进入API匹配阶段，流向原有的Cluster。

## TrafficRate
路由流量的比例，例如设置为50，那么50%的流量会根据`RoutingStrategy`进行路由。

## Status
路由的状态，只有`UP`状态才会生效。
Routing
------
A Routing represents a router. Through routing, we can implement AB test, traffic direction and other advanced features.

# Routing Attributes
## ID
Unique Identifier

## Name
Routing Name

## clusterID
The cluster to which the routing traffic goes.

## apiID
The API for which the routing is

## Condition (Optional)
Routing Condition. When the condition is met, Gateway executes this routing strategy. The routing condition can set the arguement expressions of `cookie`、`querystring`、`header`、`json body`,`path value`. If not set, all traffic is matched.

## RoutingStrategy
Currently support `Split`, which refers to redirecting a certain percentage of eligible requests to the target cluster and direct the rest to the API matching phase and then to the original cluster destination.

## TrafficRate
If set to 50, 50% of traffic is being routed according to `RoutingStrategy`.

## Status
Routing is valid only if status is `UP`.
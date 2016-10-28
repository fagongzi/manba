API
-----------
API is the core concept in gateway. You can use admin to manage your APIs. 

# API fields
* URL
  URL is a regex pattern for match request url. If a origin request url matches this value, proxy dispatch request to nodes which is defined in this api.

* Nodes
  API nodes is a list infomation. Every Node has 4 attrbutes: cluster, attrbute name, rewrite. Proxy will dispatch origin request to these nodes, and wait for all response, than merge to response to client.

  * Cluster (required)
    Which cluster to send. Proxy will use a balancer to select a backend server from this cluster to send request.

  * Attrbute name (required when nodes size > 1)
    This value is used for merge action. Example, we have 2 nodes, one is set to `base` and responsed value is `{"name":"user1"}`, one is set to `account` and responsed value is `{"money":1000}`, after merge, result is `{"base": {"name":"user1"}, "account": {"money":1000}}`.
    
  * Rewrite (optional)
    Used for you want to rewite origin url to your wanted. It usually work together with **URL** attrbute. In actual, we need use proxy for a old system, but the old system's API is design not restful friendly. In this scenes, we want to provide a beatful API design to other user. The URL rewrite is a solution. For example, a old system provide a API `/user?userId=xxx`, and we want to provide a API like this `/api/users/xxx`, you can set **Url** to `/api/users/(.+)` and set **rewite** to `/user?userId=$1`.

# What can API do

## Redefine backend server API
Through API rewite, we can redefine backend server's API. 

## Response merge
For mobile application, we can use this funcation to providing combined API for saving traffic and improving performance. 

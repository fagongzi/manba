Cluster
-------
In Gateway, cluster is a logic concept. It's a set of backend server that provide the same service.

# CRUD
## Create a cluster
A cluster contains 3 field:

* Cluster Name
  It's a uniq string at hole metedata store.

* Cluster URL Pattern
  It's a regular expressions for match http request path. e.g. /api/*.

* Cluster Load Balance
  The load balance algorithm for cluster select a backend server in the set.

You can create a cluster by admin. Once a cluster has bean created, all proxy will auto add to it's memery. You will see a info log at proxy.

## Update a cluster
You can update all field of a cluster in addition to name filed, even though the proxy is running. Once a cluster has updated, all proxy will auto update there memery. You will see a info log at proxy.

## Delete a cluster
You can delete a cluster, even though the proxy is running. Once a cluster has updated, all proxy will auto update there memery. You will see a info log at proxy.

# Bind Servers
The next step is bind backend servers in the cluster. The proxy will auto add、update、delete bind info. In gateway all metedata changed be perceived by all proxy, not need to restart the proxy node.

# Example
You have 3 servers which provide user API: `127.0.0.1:8080`, `127.0.0.1:8081`,`127.0.0.1:8082`. The user query api is: `/api/users/{id}`. You can use gateway through steps as below:

* create a cluster
  Cluster Name field set to `app`, Cluster URL Pattern field set to `/api/users/*`, Cluster Load Balance set to roundrobin.
  
* create 3 servers
  use above 3 address.

* bind 3 servers with cluster

Than you can access your proxy url to access backend servers: `http://127.0.0.1/api/users/xxx`, try it!
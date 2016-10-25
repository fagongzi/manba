Server
------
In Gateway, server is a real backend server that provide API implemention.

# Server fields
* Server Addr
  The backend server's IP and Port. The same IP And Port identified a uniq backend server.

* Server Check URL
  The URL for server heath check.

* Server Check URL Responsed Body
  The check url expect response value, if not set, proxy only check http status code is 200.

* Server Check Timeout
  Timeout of backend server response.

* Server Check Duration
  Duration of heath check.

* Server Max QPS
  Used for rate limiting, over this value, proxy will reject request.

* Server Half To Open Seconds
  It's a circuit breaker attrbutes, how many seconds that half-open status convert to open status.

* Server Half Traffic Rate
  It's a circuit breaker attrbutes, at half-open status, how many rate of traffic will passed from proxy, other will reject.

* Server To Close Count
  It's a circuit breaker attrbutes, number of continuous occur error. Over this value will convert to closed status, in closed status proxy will reject all request.

# CRUD
# Create
You can create a server use admin. Once a server created, all proxy will add it to check tasks, after heath, this server will be moved to proxy's available servers list.

# Update
You can update server's infomation at admin system. Once server has been updated, all proxy will update there memory immidately. 

# Delete
You can delete a server at admin system. Once server has been deleted, all proxy will delete it immidately. 

# Bind
Server will not received traffic until it binded a cluster. So you must bind server with a cluster.

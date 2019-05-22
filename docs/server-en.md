Server
------
In Gateway, a server refers to a real backend server.

# Server Attributes
## ID
Unique Identifier

## Addr
Format: "IP:PORT"

## Protocol
API Protocol. Currently only support HTTP

## Weight
Valid only if the load balance strategy is Weighted Round Robin

## MaxQPS
Maximum QPS supported by server. Used to Control Traffic. Gateway uses the Token Bucket Algorithm, restricting traffic by MaxQPS, thus protecting backend servers from overload.

## HealthCheck (Optional)
Health check mechanism, currently supporting HTTP check, response status code and response body. If not set, the server's health check becomes external responsibility and Gateway always assumes that this server is healthy.

## CircuitBreaker (Optional)
Backend server circuit break status:

* Open

  Normal. All traffic in. When Gateway find the failed requests to all requests ratio reach a certain threshold, CircuitBreaker switches from Open to Close.

* Half

  Attempt to recover. Gateway tries to direct a certain percentage of traffic to the server and observe the result. If the expectation is met, CircuitBreaker switches to Open. If not, Close.

* Close

  Gateway does not direct any traffic to this backend server. When the time threshold is reached, Gateway automatically tries to recover by switching to Half.
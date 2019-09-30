## Optimization
There are some places where the Gateway Gateway needs to be optimized to achieve better performance. Here are some suggestions for optimizing the configuration.

### Docker
When using docker to start the `Proxy` image, since the default image `somaxconn` is the default value of `128`, the docker starts with the following parameters:
```bash
Docker run -d --rm --sysctl net.core.somaxconn=The value you expect fagongzi/proxy
```

### Keepalive
Proxy defaults to enabling Keepalive to maintain links to backend services. There are 2 important startup parameters:
* limit-conn-keepalive link maximum keepalive time
* limit-conn-idle The maximum idle time of the link. After this time, the Proxy will actively close the connection.

Recommended configuration: `Keepalive time of backend service> limit-conn-keepalive > 2*limit-conn-idle`

### limit-dispatch
This parameter is used to set the number of working coroutine pools for the ʻaggregation` request. When there is an 'aggregation' request in the system, this value can be appropriately adjusted to improve the throughput of the `aggregation` request.

### limit-copy
This parameter is used to set the number of working coordination pools of `Copy traffic` traffic. When there is a scene of `Copy traffic` traffic in the system, this value can be adjusted appropriately to improve the throughput of `Copy traffic`.

### limit-conn
This parameter is used to control the maximum number of links that `Proxy` can establish with each backend. If it finds that the backend of a backend server is very high and the delay is large, it may take time to wait for the operation of the idle link. Increase this value.

### limit-caching
This parameter controls how much memory a single `Proxy` uses for caching. If the API in `Proxy` has caching enabled, then this parameter can be used to control the cache size.
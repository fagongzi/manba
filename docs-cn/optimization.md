## Optimization
Gateway网关在使用的时候有一些地方需要做一些优化，以便达到更好的性能。这里给出一些优化配置的意见。

### Docker
在使用docker启动`Proxy`镜像的时候，由于默认的镜像的`somaxconn`是默认值`128`，docker 启动的时候加上如下参数：
```bash
docker run -d --rm --sysctl net.core.somaxconn=你期望的值 fagongzi/proxy
```

### Keepalive
Proxy默认回启用Keepalive来保持和后端服务的链接，这里有2个重要的启动参数：
* limit-conn-keepalive 链接最大Keepalive保持的时间
* limit-conn-idle      链接最大的空闲时间，超过这个时间，Proxy会主动关闭这个连接

建议配置: `后端服务的Keepalive时间 > limit-conn-keepalive > 2*limit-conn-idle`

### limit-dispatch
这个参数是用来设置`聚合`请求的工作协程池的个数，当系统中有`聚合`请求，可以适当调大这个值来提升`聚合`请求的吞吐。

### limit-copy
这个参数是用来设置`Copy流量`流量的工作协程池的个数，当系统中有`Copy流量`流量的场景，可以适当调大这个值来提升`Copy流量`的吞吐。

### limit-conn
这个参数是用来控制`Proxy`与每个后端建立的最大链接数，如果发现某个后端Server的并发非常高，延迟较大，有可能时间花费了等待空闲链接的操作上，可以适当调大这个值。

### limit-caching
这个参数控制单个`Proxy`使用多少内存来做缓存，如果`Proxy`中的API开启了缓存，那么可以通过这个参数控制缓存大小。
Server
------
在Gateway中，一个Server对应一个真实存在的后端Server。

# Server 属性
* Server Addr
  Server Addr由后端Server的`IP:PORT`组成。

* Server Check URL
  健康检查后端URL

* Server Check URL Responsed Body
  健康检查期望返回的HTTP Body字串，如果不设置，仅仅检查返回状态码为`200`

* Server Check Timeout
  健康检查单次请求超时时间，超出这个时间后端Server不响应，这个后端Server会被标记为不可用。知道下次健康检查通过，会被重新标记为可用。

* Server Check Duration
  健康检查间隔

* Server Max QPS
  后端Server的QPS，超过这个QPS的请求会被Proxy直接拒绝。用于流控。

* Server Half To Open Seconds
  熔断机制中从半打开状态转换到打开状态的时间，单位秒。

* Server Half Traffic Rate
  熔断机制中，处于半打开状态的时候，多少比例的流量请求会被正常转发，剩余的流量直接拒绝。

* Server To Close Count
  熔断机制中，设置连续失败多少次，熔断状态从正常转换到关闭。

# 增删改查
对于Server的增删改查操作，所有的Proxy都会通过Etcd的watch机制自动感知，并且实时生效。**注意，当一个Server和Cluster绑定时，不能被删除**

# 绑定
在Gateway中，Server并不能直接接受流量，必须和Cluster绑定才能被Proxy分发流量。

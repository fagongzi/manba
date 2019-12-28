Server
------
在Manba中，一个Server对应一个真实存在的后端Server。

# Server属性
## ID
Server ID，唯一标识。

## Addr
Server地址，格式为："IP:PORT"。

## Protocol
Server的接口协议，目前支持HTTP。

## Weight
Weight 服务器的权重（当该服务器所属的集群负载方式是权重轮询时则需要配置）

## MaxQPS
Server能够支持的最大QPS，用于流控。Manba采用令牌桶算法，根据QPS限制流量，保护后端Server被压垮。

## HealthCheck（可选）
Server的健康检查机制，目前支持HTTP的协议检查，支持检查返回状态码以及返回内容。如果没有设置，认为这个Server的健康检查交给外部，Manba永久认为这个Server是健康的。

## CircuitBreaker（可选）
熔断器，设置后端Server的熔断规则。熔断器分为3个状态：

* Open

  Open状态，正常状态，Manba放入全部流量。当Manba发现失败的请求比例达到了设置的规则，熔断器会把状态切换到Close状态

* Half

  Half状态，尝试恢复的状态。在这个状态下，Manba会尝试放入一定比例的流量，然后观察这些流量的请求的情况，如果达到预期就把状态转换为Open状态，如果没有达到预期，恢复成Close状态

* Close

  Close状态，在这个状态下，Manba禁止任何流量进入这个后端Server，在达到指定的阈值时间后，Manba自动尝试切换到Half状态，尝试恢复。

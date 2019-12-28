搭建Manba环境
------------------------
这个章节帮助你搭建Manba环境

# 准备
## Etcd
Manba目前支持Etcd作为元数据区的存储，所以需要一个Etcd环境，参考：[etcd environment](https://github.com/coreos/etcd)


## Golang
如果你希望从源码编译Manba，你需要一个[golang 环境](https://github.com/golang/go)，必须使用`1.11`及以上的版本。

# 从源码编译
- 使用Makefile脚本

  以下命令默认在项目根目录（即`$GOPATH/src/github.com/fagongzi/manba`）目录下执行。

  - 编译适合当前系统的二进制文件

  ```bash
  make release_version='version string'
  ```

  - 指定编译的二进制文件类型

  ```bash
  # Linux
  make release release_version='version string'

  # Darwin(mac osx)
  make release_darwin release_version='version string'
  ```

  - 打包为docker镜像

  ```bash
  make docker release_version='version string'
  ```

  - 打包为docker镜像，且定制镜像内容

  ```bash
  # for demo, including etcd, proxy, apiserver, ui
  make docker release_version='version string'

  # only proxy
  make docker release_version='version string' with=proxy

  # only etcd
  make docker release_version='version string' with=etcd

  # apiserver with ui
  make docker release_version='version string' with=apiserver
  ```

  - 更多使用说明

  ```bash
  make help
  ```

# Manba组件

Manba运行环境包含2个组件：`ApiServer` 和 `Proxy`

* ApiServer

  可执行文件`manba-apiserver`，对外提供API管理元数据。

* Proxy

  可执行文件`manba-proxy`，Proxy是一个无状态的API代理，提供给终端用户直接访问。

## ApiServer
ApiServer对外提供GRPC的服务，用来管理Manba的元数据。

```bash
$ ./manba-apiserver --help
Usage of ./manba-apiserver:
  -addr string
    	Addr: client entrypoint (default "127.0.0.1:9091")
  -addr-store string
    	Addr: store address (default "etcd://127.0.0.1:2379")
  -crash string
    	The crash log file. (default "./crash.log")
  -discovery
    	Publish apiserver service via discovery.
  -log-file string
    	The external log file. Default log to console.
  -log-level string
    	The log level, default is info (default "info")
  -namespace string
    	The namespace to isolation the environment. (default "dev")
  -publish-lease int
    	Publish service lease seconds (default 10)
  -publish-timeout int
    	Publish service timeout seconds (default 30)
  -service-prefix string
    	The prefix for service name. (default "/services")
```

`discovery`参数用来是否使用服务发现的方式发布ApiServer提供的对外接口
`namespace`参数用来隔离多个环境，这个配置需要和对应的`Proxy`的`namespace`一致

## proxy
Proxy是内部所有API的统一对外入口，也就是API统一接入层。

```bash
$ ./manba-proxy --help
Usage of ./manba-proxy:
  -addr string
    	Addr: http request entrypoint (default "127.0.0.1:80")
  -addr-pprof string
    	Addr: pprof addr
  -addr-rpc string
    	Addr: manager request entrypoint (default "127.0.0.1:9091")
  -addr-store string
    	Addr: store of meta data, support etcd (default "etcd://127.0.0.1:2379")
  -crash string
    	The crash log file. (default "./crash.log")
  -filter value
    	Plugin(Filter): format is <filter name>[:plugin file path][:plugin config file path]
  -limit-body int
    	Limit(MB): MB for body size (default 10)
  -limit-buf-read int
    	Limit(bytes): Bytes for read buffer size (default 2048)
  -limit-buf-write int
    	Limit(bytes): Bytes for write buffer size (default 1024)
  -limit-conn int
    	Limit(count): Count of connection per backend server (default 64)
  -limit-conn-idle int
    	Limit(sec): Idle for backend server connections (default 30)
  -limit-conn-keepalive int
    	Limit(sec): Keepalive for backend server connections (default 60)
  -limit-heathcheck int
    	Limit: Count of heath check worker (default 1)
  -limit-heathcheck-interval int
    	Limit(sec): Interval for heath check (default 60)
  -limit-timeout-read int
    	Limit(sec): Timeout for read from backend servers (default 30)
  -limit-timeout-write int
    	Limit(sec): Timeout for write to backend servers (default 30)
  -log-file string
    	The external log file. Default log to console.
  -log-level string
    	The log level, default is info (default "info")
  -namespace string
    	The namespace to isolation the environment. (default "dev")
  -ttl-proxy int
    	TTL(secs): proxy (default 10)
  -version
      Show version info
```

`namespace`参数用来隔离多个环境，这个配置需要和对应的`ApiServer`的`namespace`一致

# 运行环境
我们以三台etcd、一台ApiServer，三台Proxy的环境为例

## 环境信息

|组件|环境|
| -------------|:-------------:|
|etcd集群环境|192.168.1.100,192.168.1.101,192.168.1.102|
|Proxy|192.168.1.200,192.168.1.201,192.168.1.202|
|ApiServer|192.168.1.203|

## 启动Proxy
```bash
./manba-proxy --addr=192.168.1.200:80 --addr-rpc=192.168.1.200:9091 --addr-store=etcd://192.168.1.100:2379,192.168.1.101:2379,192.168.1.102:2379 --namespace=test
```

```bash
./manba-proxy --addr=192.168.1.201:80 --addr-rpc=192.168.1.201:9091 --addr-store=etcd://192.168.1.100:2379,192.168.1.101:2379,192.168.1.102:2379 --namespace=test
```

```bash
./manba-proxy --addr=192.168.1.202:80 --addr-rpc=192.168.1.202:9091 --addr-store=etcd://192.168.1.100:2379,192.168.1.101:2379,192.168.1.102:2379 --namespace=test
```

用户的API接入地址可以为：192.168.1.201:80、192.168.1.201:80、192.168.1.202:80其中任意一个

## 启动ApiServer
```bash
./manba-apiserver --addr=192.168.1.203:9091 --addr-store=etcd://192.168.1.100:2379,192.168.1.101:2379,192.168.1.102:2379 --discovery --namespace=test
```

## 调用ApiServer创建元信息
[Gateway Restful API](./restful.md)

[Gateway grpc客户端例子](../examples)

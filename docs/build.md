Gateway Environment Setup
------------------------
This chapter aims to help you set up Gateway environment.

# Preparation
## Etcd
Gateway currently supports Etcd as the storage of metadata. To set up etcd, please refer to [etcd environment](https://github.com/coreos/etcd)


## Golang
If you would like to compile Gateway from source code, you need a [Golang environment](https://github.com/golang/go). Go version `1.11` and above is required.

# Compiling from Source Code
- Makefile

  The following commands are executed under `$GOPATH/src/github.com/fagongzi/gateway`.

  - Compiling binary file for the current OS.

  ```bash
  make release_version='version string'
  ```

  - Compiling binary file for a specific OS

  ```bash
  # Linux
  make release release_version='version string'

  # Darwin(Mac OS X)
  make release_darwin release_version='version string'
  ```

  - Packaging project into a Docker image

  ```bash
  make docker release_version='version string'
  ```

  - Packing project into a Docker image with customized content

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

  - For more information on compiling

  ```bash
  make help
  ```

# Gateway Component
Gateway has two components: `ApiServer` and `Proxy`.

* ApiServer
  ApiServer provides APIs to manage metadata.

* Proxy
  Proxy is a stateless API proxy which provides direct access to clients.

## ApiServer
ApiServer provides GRPC service to manage Gateway metadata.

```bash
$ ./apiserver --help
Usage of ./apiserver:
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

`discovery` option is used to determine whether to use service discovery to publish external APIs provided by ApiServer.
`namespace` option is used to isolate multiple environments. It has to be consistent with `namespace` in `Proxy`.


## proxy
Proxy is the unified entrance of all internal APIs, which is the API access layer.

```bash
$ ./proxy --help
Usage of ./proxy:
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

`namespace` option is used to isolate multiple environments. It has to be consistent with `namespace` in `ApiServer`.

# Running Environment
We use 3 etcd servers, 1 ApiServer server, and 3 Proxy servers as an example.

## Info

|Component|IP|
| -------------|:-------------:|
|etcd cluster|192.168.1.100,192.168.1.101,192.168.1.102|
|Proxy|192.168.1.200,192.168.1.201,192.168.1.202|
|ApiServer|192.168.1.203|

## Starting Proxy
```bash
./proxy --addr=192.168.1.200:80 --addr-rpc=192.168.1.200:9091 --addr-store=etcd://192.168.1.100:2379,192.168.1.101:2379,192.168.1.102:2379 --namespace=test
```

```bash
./proxy --addr=192.168.1.201:80 --addr-rpc=192.168.1.201:9091 --addr-store=etcd://192.168.1.100:2379,192.168.1.101:2379,192.168.1.102:2379 --namespace=test
```

```bash
./proxy --addr=192.168.1.202:80 --addr-rpc=192.168.1.202:9091 --addr-store=etcd://192.168.1.100:2379,192.168.1.101:2379,192.168.1.102:2379 --namespace=test
```

API addresses available to users: 192.168.1.201:80, 192.168.1.201:80, 192.168.1.202:80

## Starting ApiServer
```bash
./apiserver --addr=192.168.1.203:9091 --addr-store=etcd://192.168.1.100:2379,192.168.1.101:2379,192.168.1.102:2379 --discovery --namespace=test
```

## Use ApiServer to create metadata
[Gateway Restful API](./restful.md)

[Gateway grpc client example](../examples)

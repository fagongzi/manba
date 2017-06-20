Build and run Gateway
------------------------
This section help you build gateway environment.

# prepare
## Etcd
Currently, Gateway use etcd store it's mete data, so you need a [etcd environment](https://github.com/coreos/etcd).

## Consul
Currently, Gateway use consul store it's mete data, so you need a [consul environment](https://github.com/hashicorp/consul).

## Golang
If you want to build gateway with source, you need a [golang environment](https://github.com/golang/go). 

# Build with source 
Build gateway use commands as below：

```bash
cd $GOPATH/src/github.com/fagongzi/gateway/cmd/proxy
go build proxy.go

cd $GOPATH/src/github.com/fagongzi/gateway/cmd/admin
go build admin.go
```

# Run Gateway
Gateway runtime has 2 components: admin & proxy. Admin is mete data manager system, proxy is a stateless http proxy.

## Run admin
Admin is web system, provide JSON restful API, web resource is in `$$GOPATH/src/github.com/fagongzi/gateway/cmd/admin/public` dir. You can get help info use:

```bash
$ ./admin --help
Usage of xxx/src/github.com/fagongzi/gateway/cmd/admin/admin:
  -addr string
        listen addr.(e.g. ip:port) (default ":8080")
  -cpus int
        use cpu nums (default 1)
  -registry-addr string
        registry address. (default "[ectd|consul]://127.0.0.1:8500")
  -prefix string
        node prefix. (default "/dev")
  -pwd string
        admin user pwd (default "admin")
  -user string
        admin user name (default "admin")
```

Than you can run admin use:

```bash
./admin --addr=:8080  --etcd-addr=ectd://etcdIP:etcdPort --prefix=dev 
```

It listen at 8080 port, you can you your web browser access `http://127.0.0.1:8080`, the input the user name and password to access admin system.

## Run proxy
You can get help info use:

```bash
$ ./proxy --help
Usage of xxx/src/github.com/fagongzi/gateway/cmd/proxy/proxy:
  -config string
        config file
  -cpus int
        use cpu nums (default 1)
  -log-file string
        which file to record log, if not set stdout to use.
  -log-level string
        log level. (default "info")
```

Proxy use a json config file like this:

```json
{
    "addr": ":80", 
    "mgrAddr": ":8081",
    "registryAddr": [
        "ectd://127.0.0.1:2379"
    ],
    "prefix": "/dev",
    "filers": [
        "analysis",
        "rate-limiting",
        "circuit-breake",
        "http-access",
        "head",
        "xforward"
    ],
    "maxConns": 512,
    "maxConnDuration": 10,
    "maxIdleConnDuration": 10,
    "readBufferSize": 4096,
    "writeBufferSize": 4096,
    "readTimeout": 30,
    "writeTimeout": 30,
    "maxResponseBodySize": 1048576,

    "enablePPROF": false,
    "pprofAddr": ""
}
```

Note: Admin and proxy must use same ectd address and ectd prefix.

Run proxy:

```bash
./proxy --cpus=number of you cpu core ---config ./proxy.json --log-file ./proxy.log --log-level=info
```

Than you can see proxy start at 80 port. And load mete data from ectd. At first time, there will be have some warn message, ingore these, because has no mete data in ectd(consul). 
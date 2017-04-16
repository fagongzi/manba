Build and run Gateway
------------------------
This section help you build gateway environment.

# prepare
## Etcd
Currently, Gateway use etcd store it's mete data, so you need a [etcd environment](https://github.com/coreos/etcd).

## Golang
If you want to build gateway with source, you need a [golang environment](https://github.com/golang/go). Otherwise, you can get last binary package from [here](http://7xtbpp.com1.z0.glb.clouddn.com/gateway-linux64.tar.gz)

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
  -etcd-addr string
        etcd address, use ',' to splite. (default "http://192.168.159.130:2379")
  -etcd-prefix string
        etcd node prefix. (default "/gateway")
  -pwd string
        admin user pwd (default "admin")
  -user string
        admin user name (default "admin")
```

Than you can run admin use:

```bash
./admin --addr=:8080  --etcd-addr=http://etcdIP:etcdPort --ectd-prefix=dev 
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
    "etcdAddrs": [
        "http://127.0.0.1:2379"
    ],
    "etcdPrefix": "/dev",
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

Than you can see proxy start at 80 port. And load mete data from ectd. At first time, there will be have some warn message, ingore these, because ectd has no mete data in ectd. 
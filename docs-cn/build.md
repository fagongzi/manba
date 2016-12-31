搭建Gateway环境
------------------------
这个章节帮助你搭建Gateway环境

# 准备
## Etcd
Gateway目前支持Etcd作为元数据区的存储，所以需要一个Etcd环境，参考：[etcd environment](https://github.com/coreos/etcd)

## Golang
如果你希望从源码编译Gateway，你需要一个[golang 环境](https://github.com/golang/go)，建议使用`1.7`的版本。

# 从源码编译
执行如下命令：

```bash
cd $GOPATH/src/github.com/fagongzi/gateway/cmd/proxy
go build proxy.go

cd $GOPATH/src/github.com/fagongzi/gateway/cmd/admin
go build admin.go
```

# 运行 Gateway
Gateway运行环境包含2个组件：`Admin` 和 `Proxy`

* Admin
  后台管理系统，管理Gateway的元数据。

* Proxy
  Proxy是一个无状态的API代理，提供给终端用户直接访问。

## 运行 Admin
Admin是一个WEB应用，提供`JSON restful API`，web的静态资源在`$GOPATH/src/github.com/fagongzi/gateway/cmd/admin/public`，静态资源需要和Admin二进制程序处于一个目录。

```bash
$ ./admin --help
Usage of $GOPATH/src/github.com/fagongzi/gateway/cmd/admin/admin:
  -addr string
        listen addr.(e.g. ip:port) (default ":8080")
  -cpus int
        use cpu nums (default 1)
  -etcd-addr string
        etcd address, use ',' to splite. (default "http://127.0.0.1:2379")
  -etcd-prefix string
        etcd node prefix. (default "/dev")
  -pwd string
        admin user pwd (default "admin")
  -user string
        admin user name (default "admin")
```

执行一下命令启动Admin:

```bash
./admin --addr=:8080  --etcd-addr=http://etcdIP:etcdPort --ectd-prefix=dev 
```
启动后，Admin监听8080端口，你可以再浏览器上访问`http://127.0.0.1:8080`，默认的用户名和密码是`admin/admin`

可以使用`etcd-prefix`参数隔离多个Gateway环境。

## 运行 proxy
You can get help info use:

```bash
$ ./proxy --help
Usage of $GOPATH/src/github.com/fagongzi/gateway/cmd/proxy/proxy:
  -config string
        config file
  -cpus int
        use cpu nums (default 1)
  -log-file string
        which file to record log, if not set stdout to use.
  -log-level string
        log level. (default "info")
```

Proxy启动以来一个JSON的配置文件，这个配置文件样例可以在`$GOPATH/src/github.com/fagongzi/gateway/cmd/proxy`目录下找到。

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

注意： Admin和Proxy必须使用统一的`ectd-prefix`参数。

运行 proxy:

```bash
./proxy --cpus=number of you cpu core ---config ./proxy.json --log-file ./proxy.log --log-level=info
```

启动后，Proxy监听80端口，并且加载Etcd中的元数据，同时监听Ectd元数据的变化，并且实时更新。在第一次启动的时候，会有一些`Warn`的信息提示，这个是因为一开始Ectd中没有相关数据造成的，可以忽略。
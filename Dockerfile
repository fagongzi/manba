FROM golang

RUN mkdir -p /app/gateway
RUN mkdir -p /app/gateway/plugins
RUN mkdir -p /go/src/github.com/fagongzi/gateway 

COPY ./ /go/src/github.com/fagongzi/gateway

RUN ETCD_VER=v3.0.14 \
    && DOWNLOAD_URL=https://github.com/coreos/etcd/releases/download \
    && curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz \
    && tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /app/gateway --strip-components=1

RUN cd /go/src/github.com/fagongzi/gateway/cmd/admin \
    && go build admin.go \
    && mv ./admin /app/gateway \
    && mv ./public /app/gateway

RUN cd /go/src/github.com/fagongzi/gateway/cmd/proxy \
    && go build proxy.go \
    && mv ./proxy /app/gateway \
    && mv ./config_etcd.json  /app/gateway

COPY ./entrypoint.sh /app/gateway
RUN chmod +x /app/gateway/entrypoint.sh

EXPOSE 80
EXPOSE 8080
EXPOSE 8081

WORKDIR /app/gateway

ENTRYPOINT ./entrypoint.sh

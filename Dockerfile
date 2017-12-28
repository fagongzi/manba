FROM golang

RUN mkdir -p /app/gateway
RUN mkdir -p /app/gateway/plugins
RUN mkdir -p /go/src/github.com/fagongzi/gateway 

COPY ./ /go/src/github.com/fagongzi/gateway

RUN ETCD_VER=v3.0.14 \
    && DOWNLOAD_URL=https://github.com/coreos/etcd/releases/download \
    && curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz \
    && tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /app/gateway --strip-components=1

RUN cd /go/src/github.com/fagongzi/gateway/cmd/api \
    && go build -o apiserver -ldflags "-w -s" ./... \
    && mv ./apiserver /app/gateway

RUN cd /go/src/github.com/fagongzi/gateway/cmd/proxy \
    && go build -ldflags "-w -s" proxy.go \
    && mv ./proxy /app/gateway

COPY ./entrypoint.sh /app/gateway
RUN chmod +x /app/gateway/entrypoint.sh

ENV GATEWAY_LOG_LEVEL=info

EXPOSE 2379
EXPOSE 80
EXPOSE 9092

WORKDIR /app/gateway

ENTRYPOINT ./entrypoint.sh

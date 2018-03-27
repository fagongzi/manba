FROM alpine:latest

RUN mkdir -p /app/gateway
RUN mkdir -p /app/gateway/plugins

ADD ./dist/proxy /app/gateway
ADD ./dist/apiserver /app/gateway
ADD ./dist/etcd /app/gateway
ADD ./entrypoint.sh /app/gateway

RUN chmod +x /app/gateway/entrypoint.sh

ENV GATEWAY_LOG_LEVEL=info

EXPOSE 80 2379 9092

WORKDIR /app/gateway
ENTRYPOINT ["/bin/sh", "./entrypoint.sh"]

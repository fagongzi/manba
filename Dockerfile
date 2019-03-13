FROM alpine:latest

ARG APP_ROOT=/app/gateway
ARG EXEC_NAME=proxy
ARG UID=2019
ARG CMD_NAME=demo
ENV CURRENT_EXEC_PATH=${APP_ROOT}/${EXEC_NAME}

WORKDIR ${APP_ROOT}

ADD dist ${APP_ROOT}

# Alpine Linux doesn't use pam, which means that there is no /etc/nsswitch.conf,
# but Golang relies on /etc/nsswitch.conf to check the order of DNS resolving
# (see https://github.com/golang/go/commit/9dee7771f561cf6aee081c0af6658cc81fac3918)
# To fix this we just create /etc/nsswitch.conf and add the following line:
# hosts: files mdns4_minimal [NOTFOUND=return] dns mdns4

RUN MAIN_VERSION=$(cat /etc/alpine-release | cut -d '.' -f 0-2) \
    && mv /etc/apk/repositories /etc/apk/repositories-bak \
    && { \
        echo "https://mirrors.aliyun.com/alpine/v${MAIN_VERSION}/main"; \
        echo "https://mirrors.aliyun.com/alpine/v${MAIN_VERSION}/community"; \
    } >> /etc/apk/repositories \
    && apk add --update --no-cache libcap \
    && addgroup -g ${UID} -S gateway \
    && adduser -u ${UID} -S gateway -G gateway \
    && mkdir -p ${APP_ROOT}/plugins \
    && chown -R gateway:gateway ./ \
    && if [ -e ${CURRENT_EXEC_PATH} ]; then \
         setcap CAP_NET_BIND_SERVICE=+eip ${CURRENT_EXEC_PATH}; \
       fi \
    && echo 'hosts: files mdns4_minimal [NOTFOUND=return] dns mdns4' >> /etc/nsswitch.conf \
    && echo -n ${CMD_NAME} > cmd

USER gateway

EXPOSE 80 2379 9092 9093

ENTRYPOINT ["/bin/sh", "./entrypoint.sh"]

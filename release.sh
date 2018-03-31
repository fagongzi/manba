#!/bin/bash

HOME=`dirname $0`
DIST=$HOME/dist
BIN_API=$HOME/cmd/api
BIN_PROXY=$HOME/cmd/proxy

VERSION_PATH=`echo $(realpath $HOME) | sed -e "s;${GOPATH}/src/;;g"`/pkg/util

LD_GIT_COMMIT="-X '${VERSION_PATH}.GitCommit=`git rev-parse --short HEAD`'"
LD_BUILD_TIME="-X '${VERSION_PATH}.BuildTime=`date +%FT%T%z`'"
LD_GO_VERSION="-X '${VERSION_PATH}.GoVersion=`go version`'"
LD_FLAGS="${LD_GIT_COMMIT} ${LD_BUILD_TIME} ${LD_GO_VERSION} -w -s"

prepare() {
    echo "start prepare release"
    rm -rf $DIST
    mkdir $DIST
    echo "complete prepare release"
}

download_etcd() {
    ETCD_VER=v3.0.14 \
    && DOWNLOAD_URL=https://github.com/coreos/etcd/releases/download \
    && curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz \
    && tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C $DIST --strip-components=1
}

build_bin() {
    echo "start build binary"
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o apiserver -ldflags "${LD_FLAGS}" $BIN_API/...
    mv apiserver $DIST

    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o proxy -ldflags "${LD_FLAGS}" $BIN_PROXY/...
    mv proxy $DIST
    echo "complete build binary"
}

build_docker() {
    echo "start build docker image, version is $1"
    docker build -t fagongzi/gateway:$1 -f Dockerfile .
    docker build -t fagongzi/proxy:$1 -f Dockerfile-proxy .
    docker build -t fagongzi/apiserver:$1 -f Dockerfile-apiserver .
    echo "complete build docker"
}

clean() {
    echo "start clean build"
    rm -rf $DIST
    echo "complete clean build"
}

prepare
build_bin
download_etcd
build_docker $1
clean
echo "All completed"

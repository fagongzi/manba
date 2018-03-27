#!/bin/bash

HOME=`dirname $0`
DIST=$HOME/dist
BIN_API=$HOME/cmd/api
BIN_PROXY=$HOME/cmd/proxy

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
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o apiserver -ldflags "-w -s" $BIN_API/...
    mv apiserver $DIST

    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o proxy -ldflags "-w -s" $BIN_PROXY/...
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
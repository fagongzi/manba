#!/bin/bash

start_etcd() {
    ./etcd &
}

start_apiserver() {
    ./apiserver --addr=:9092 --discovery &
}

start_proxy() {
    ./proxy --addr=:80 --log-level=$GATEWAY_LOG_LEVEL
}

start_etcd
sleep 3
start_apiserver
sleep 1
start_proxy


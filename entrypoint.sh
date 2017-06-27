#!/bin/bash

start_etcd() {
    ./etcd &
}

start_admin() {
    ./admin --registry-addr etcd://127.0.0.1:2379 --addr=127.0.0.1:8080 &
}

start_proxy() {
    ./proxy --config ./config.json
}

start_etcd
sleep 3
start_admin
sleep 1
start_proxy


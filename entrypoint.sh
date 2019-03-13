#!/bin/sh

set -e

start_etcd() {
    ./etcd $ETCD_OPTS &
}

start_apiserver() {
    ./apiserver --addr=:9092 --addr-http=:9093 --discovery $API_SERVER_OPTS &
}

PARAM=$1
CMD=`cat cmd`
if [ "$PARAM" = "" ]
then
    PARAM=${CMD}
fi

DEFAULT_EXEC="./proxy --addr=:80 --log-level=$GATEWAY_LOG_LEVEL $GW_PROXY_OPTS"
if [ "${PARAM}" = 'demo' ]
then
    start_etcd
    sleep 3
    start_apiserver
    sleep 1
    EXEC=$DEFAULT_EXEC
fi

if [ "${PARAM}" = 'proxy' ]
then
    EXEC=$DEFAULT_EXEC
fi

if [ "${PARAM}" = 'apiserver' ]
then
    EXEC="./apiserver --addr=:9092 --addr-http=:9093 --discovery $API_SERVER_OPTS"
fi

if [ "${PARAM}" = 'etcd' ]
then
    EXEC="./etcd $ETCD_OPTS"
fi

if [ ! -z "${PARAM}" ] && [ -z "$EXEC" ]
then
    EXEC="${PARAM}"
fi

exec $EXEC

#!/bin/sh

set -e

start_etcd() {
    etcd $ETCD_OPTS &
}

DEFAULT_IP="0.0.0.0"

start_apiserver() {
    apiserver --addr=${DEFAULT_IP}:9092 --addr-http=${DEFAULT_IP}:9093 --discovery $API_SERVER_OPTS &
}

INPUT_CMD=$@
CMD=`cat cmd`
if [ "$INPUT_CMD" = "" ]
then
    INPUT_CMD=${CMD}
fi

DEFAULT_EXEC="proxy --addr=${DEFAULT_IP}:80 --log-level=$GATEWAY_LOG_LEVEL $GW_PROXY_OPTS"
if [ "${INPUT_CMD}" = 'demo' ]
then
    start_etcd
    sleep 3
    start_apiserver
    sleep 1
    EXEC=$DEFAULT_EXEC
fi

if [ "${INPUT_CMD}" = 'proxy' ]
then
    EXEC=$DEFAULT_EXEC
fi

if [ "${INPUT_CMD}" = 'apiserver' ]
then
    EXEC="apiserver --addr=${DEFAULT_IP}:9092 --addr-http=${DEFAULT_IP}:9093 --discovery $API_SERVER_OPTS"
fi

if [ "${INPUT_CMD}" = 'etcd' ]
then
    EXEC="etcd $ETCD_OPTS"
fi

if [ ! -z "${INPUT_CMD}" ] && [ -z "$EXEC" ]
then
    EXEC=${INPUT_CMD}
fi

exec $EXEC

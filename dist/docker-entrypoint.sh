#!/usr/bin/env bash

set -e

RUN_CONF_DIR=
case "`uname`" in
    Darwin)
        RUN_CONF_DIR="${RUN_CONF_DIR:-/etc/defaults}"
        ;;
    Linux)
        RUN_CONF_DIR="${RUN_CONF_DIR:-/etc/default}"
        ;;
esac

# Read an optional runtime configuration file
if [ -z "$RUN_CONF" ]; then
    RUN_CONF="${RUN_CONF_DIR}/autoscale.conf"
fi
if [ -r "$RUN_CONF" ]; then
    . "$RUN_CONF"
fi

PRG_ARGS="${PRG_ARGS} --name=${AUTOSCALE_NAME}"
PRG_ARGS="${PRG_ARGS} --threshold=${AUTOSCALE_THRESHOLD}"
PRG_ARGS="${PRG_ARGS} --max=${AUTOSCALE_MAX}"
PRG_ARGS="${PRG_ARGS} --db-dir=/data/db"

if [ -z "$RABBITMQ_URI" ]; then
    RABBITMQ_VHOST=${RABBITMQ_VHOST:-/}
    RABBITMQ_URI=amqp://${RABBITMQ_USER}:${RABBITMQ_PASS}@${RABBITMQ_SERVICE_HOST}:${RABBITMQ_SERVICE_PORT}/${RABBITMQ_VHOST}
fi

if [ -n "$RABBITMQ_URI" ]; then
    PRG_ARGS="${PRG_ARGS} --amqp-uri='${RABBITMQ_URI}'"
fi

if [ -n "$RABBITMQ_QUEUE" ]; then
    PRG_ARGS="${PRG_ARGS} --amqp-queue=${RABBITMQ_QUEUE}"
fi

if [ -z "$KUBERNETES_SERVICE_URL" ]; then
    KUBERNETES_SERVICE_HOST=${KUBERNETES_SERVICE_HOST:-127.0.0.1}
    KUBERNETES_SERVICE_PORT=${KUBERNETES_SERVICE_HOST:-443}
    KUBERNETES_SERVICE_PROTO=${KUBERNETES_SERVICE_PROTO:-https}
    KUBERNETES_SERVICE_URL=${KUBERNETES_SERVICE_PROTO}://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT}
fi

if [ -n "$KUBERNETES_SERVICE_URL" ]; then
    PRG_ARGS="${PRG_ARGS} --api-url='${KUBERNETES_SERVICE_URL}'"
fi

if [ -n "$KUBERNETES_SERVICE_USERNAME" ]; then
    PRG_ARGS="${PRG_ARGS} --api-user='${KUBERNETES_SERVICE_USERNAME}'"
fi

if [ -n "$KUBERNETES_SERVICE_PASSWORD" ]; then
    PRG_ARGS="${PRG_ARGS} --api-passwd='${KUBERNETES_SERVICE_PASSWORD}'"
fi

if [ -f /var/run/secrets/kubernetes.io/serviceaccount/token ]; then
    PRG_ARGS="${PRG_ARGS} --api-token=/var/run/secrets/kubernetes.io/serviceaccount/token"
fi

if [ -f /var/run/secrets/kubernetes.io/serviceaccount/ca.crt ]; then
    PRG_ARGS="${PRG_ARGS} --api-cafile=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
fi

if [ -n "$KUBERNETES_SERVICE_INSECURE" ]; then
    PRG_ARGS="${PRG_ARGS} --api-insecure='${KUBERNETES_SERVICE_INSECURE}'"
fi

if [ -n "$AUTOSCALE_PERIOD" ]; then
    PRG_ARGS="${PRG_ARGS} --period=${AUTOSCALE_PERIOD}"
fi

if [ -n "$AUTOSCALE_INTERVAL" ]; then
    PRG_ARGS="${PRG_ARGS} --interval=${AUTOSCALE_INTERVAL}"
fi

if [ -n "$AUTOSCALE_STATS_COVERAGE" ]; then
    PRG_ARGS="${PRG_ARGS} --stats-coverage=${AUTOSCALE_STATS_COVERAGE}"
fi

if [ -n "$AUTOSCALE_STATS_INTERVAL" ]; then
    PRG_ARGS="${PRG_ARGS} --stats-interval=${AUTOSCALE_STATS_INTERVAL}"
fi

if [ -n "$AUTOSCALE_MIN" ]; then
    PRG_ARGS="${PRG_ARGS} --min=${AUTOSCALE_MIN}"
fi

if [ -n "$AUTOSCALE_KIND" ]; then
    PRG_ARGS="${PRG_ARGS} --kind=${AUTOSCALE_KIND}"
fi

if [ -n "$AUTOSCALE_NS" ]; then
    PRG_ARGS="${PRG_ARGS} --ns=${AUTOSCALE_NS}"
fi

if [ -n "$AUTOSCALE_DB" ]; then
    PRG_ARGS="${PRG_ARGS} --db=${AUTOSCALE_DB}"
fi

BIN_DIR=${BIN_DIR:-`pwd`}
PRG_BIN=${BIN_DIR}/autoscale

eval ${PRG_BIN} ${PRG_ARGS} 2>&1 $@

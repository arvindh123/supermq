#!/bin/bash
set -e


# Substitute env vars into rabbitmq.conf
envsubst '${SMQ_SERVICE_RABBITMQ_PORT}
    ${SMQ_SERVICE_RABBITMQ_TLS_PORT}
    ${SMQ_RABBITMQ_AMQP_TLS_PORT}
    ${SMQ_RABBITMQ_AMQP_PORT}
    ${SMQ_RABBITMQ_MQTT_PORT}
    ${SMQ_RABBITMQ_MQTT_TLS_PORT}
    ${SMQ_RABBITMQ_MQTT_WS_PATH}
    ${SMQ_RABBITMQ_MQTT_WS_PORT}
    ${SMQ_RABBITMQ_MQTT_WSS_PORT}
    ${SMQ_AUTH_GRPC_HOST}
    ${SMQ_AUTH_GRPC_PORT}
    ${SMQ_CLIENTS_GRPC_HOST}
    ${SMQ_CLIENTS_GRPC_PORT}
    ${SMQ_CHANNELS_GRPC_HOST}
    ${SMQ_CHANNELS_GRPC_PORT}
    ${SMQ_RABBITMQ_MANAGEMENT_HTTP_PORT}
    ${SMQ_SERVICE_RABBITMQ_PORT}
    ${SMQ_SERVICE_RABBITMQ_TLS_PORT}' < /etc/rabbitmq/rabbitmq.conf.template > /etc/rabbitmq/conf.d/10-defaults.conf

# If no args passed, use rabbitmq-server
if [ $# -eq 0 ]; then
  set -- rabbitmq-server
fi

exec /usr/local/bin/docker-entrypoint.sh "$@"

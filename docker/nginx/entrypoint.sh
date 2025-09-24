#!/bin/ash
# Copyright (c) Abstract Machines
# SPDX-License-Identifier: Apache-2.0

if [ -z "$SMQ_MQTT_CLUSTER" ]
then
      envsubst '${SMQ_RABBITMQ_MQTT_PORT} ${SMQ_RABBITMQ_MQTT_TLS_PORT}' < /etc/nginx/snippets/mqtt-upstream-single.conf > /etc/nginx/snippets/mqtt-upstream.conf
      envsubst '${SMQ_RABBITMQ_MQTT_WS_PORT} ${SMQ_RABBITMQ_MQTT_WS_PATH}' < /etc/nginx/snippets/mqtt-ws-upstream-single.conf > /etc/nginx/snippets/mqtt-ws-upstream.conf
else
      envsubst '${SMQ_RABBITMQ_MQTT_PORT} ${SMQ_RABBITMQ_MQTT_TLS_PORT}' < /etc/nginx/snippets/mqtt-upstream-cluster.conf > /etc/nginx/snippets/mqtt-upstream.conf
      envsubst '${SMQ_RABBITMQ_MQTT_WS_PORT}' < /etc/nginx/snippets/mqtt-ws-upstream-cluster.conf > /etc/nginx/snippets/mqtt-ws-upstream.conf
fi

envsubst '
    ${SMQ_NGINX_SERVER_NAME}
    ${SMQ_AUTH_HTTP_PORT}
    ${SMQ_DOMAINS_HTTP_PORT}
    ${SMQ_GROUPS_HTTP_PORT}
    ${SMQ_USERS_HTTP_PORT}
    ${SMQ_CLIENTS_HTTP_PORT}
    ${SMQ_CLIENTS_AUTH_HTTP_PORT}
    ${SMQ_CHANNELS_HTTP_PORT}
    ${SMQ_HTTP_ADAPTER_PORT}
    ${SMQ_NGINX_MQTT_PORT}
    ${SMQ_NGINX_MQTTS_PORT}' < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf

exec nginx -g "daemon off;"

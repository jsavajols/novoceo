#!/bin/sh
set -e

mkdir -p /mosquitto/data /mosquitto/log

cat > /tmp/mosquitto.conf << 'CONF'
listener 1883 0.0.0.0
allow_anonymous true
persistence true
persistence_location /mosquitto/data/
log_dest stdout
CONF

exec /usr/sbin/mosquitto -c /tmp/mosquitto.conf

#!/bin/bash

set -e

AGGRESSIVE_OPTS="-server -XX:+UseCompressedOops"
OTHER_OPTS="-XX:-OmitStackTraceInFastThrow"

exec java $AGGRESSIVE_OPTS $OTHER_OPTS -cp /app/riemann.jar riemann.bin start /satori-conf/rules/riemann.config

#!/bin/bash

set -e

AGGRESSIVE_OPTS="-server -XX:+UseConcMarkSweepGC -XX:+UseParNewGC -XX:+CMSParallelRemarkEnabled -XX:+AggressiveOpts -XX:+UseFastAccessorMethods -XX:+UseCompressedOops -XX:+CMSClassUnloadingEnabled"
OTHER_OPTS="-XX:-OmitStackTraceInFastThrow"

exec java $AGGRESSIVE_OPTS $OTHER_OPTS -cp /app/riemann.jar riemann.bin start /satori-conf/rules/riemann.config

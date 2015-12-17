#!/bin/bash

set -x

cd scripts/external/perf-client/
./runjava.sh com.rabbitmq.examples.PerfTest \
  --exchange perf-test \
  -uri amqp://guest:guest@localhost:${RUN_PORT} \
  --queue perf-test-transient \
  --consumers 4 \
  --producers 2 \
  --qos 20 \
  --time 20 "$@"

#!/usr/bin/env bash
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${script_dir}/lib.sh

echo "-- elasticsearch"
delete  "e2e/resources/es"

echo "-- fluentd"
delete "e2e/resources/fluentd"

echo "-- logging-pod"
delete "e2e/resources/logging-pod.yml"
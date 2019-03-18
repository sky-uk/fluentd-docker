#!/usr/bin/env bash
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${script_dir}/lib.sh

echo "-- elasticsearch"
deploy "kube-system" "k8s-app=elasticsearch" "e2e/resources/es"
if [[ $? -ne 0 ]]; then
    exit -1
fi
#
#echo "-- fluentd"
#deploy "kube-system" "k8s-app=fluentd-es" "e2e/resources/fluentd"
#if [[ $? -ne 0 ]]; then
#    exit -1
#fi
#
#echo "-- logging-pod"
#deploy "kube-system" "app=logger" "e2e/resources/logging-pod.yml"
#if [[ $? -ne 0 ]]; then
#    exit -1
#fi
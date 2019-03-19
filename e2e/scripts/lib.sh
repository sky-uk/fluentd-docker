#!/usr/bin/env bash
kind_cluster=${KIND_CLUSTER:-'es-e2e'}

function get_kubeconfig(){
    echo "--kubeconfig=$(kind get kubeconfig-path --name ${kind_cluster})"
}

function get_pods() {
    local namespace=$1
    local pod_label=$2
    local template=$3
    kubectl $(get_kubeconfig) --namespace=${namespace} get pod -l ${pod_label} -o go-template='{{ range .items }}{{ printf "%s\n" .metadata.name }}{{ end }}'
    return $?
}


function get_pods_status() {
    local namespace=$1
    local pod_label=$2
    local template=$3
    kubectl $(get_kubeconfig) --namespace=${namespace} get pod -l ${pod_label} -o go-template='{{ range .items }}{{ .metadata.name }}: {{ range .status.conditions }}{{ .type }}={{ .status }} {{ end }}{{ print "\n" }}{{ end }}'
    return $?
}

function are_pods_running() {
    local namespace=$1
    local pod_label=$2
    pods=$(get_pods ${namespace} ${pod_label})
    [[ $(echo ${pods} | wc -l) -gt 0 ]]
    return $?
}

function are_pods_ready() {
    local namespace=$1
    local pod_label=$2
    # Check statuses of pods
    are_pods_running ${namespace} ${pod_label}
    if [[ $? -eq 0 ]]; then
        pods=$(get_pods_status ${namespace} ${pod_label})
        if [[ $? -eq 0 ]]; then
            [[ $(echo ${pods} | grep "Ready=False" | wc -l) -eq 0 ]]
            return $?
        fi
    fi
    return 1
}

function wait_for_pods(){
    local namespace=$1
    local pod_label=$2
    local timeout=$3
    timer=0
    # Give kubelet a chance to register.
    echo "Waiting for ${namespace}.${pod_label} to be ready..."
    while ! are_pods_ready ${namespace} ${pod_label} ; do
        timer=$(($timer + 1))
        sleep 1
        if [[ ${timer} -gt ${timeout} ]]; then
            echo "Exit - ${namespace}.${pod_label} never went ready after ${timer}s"
            return 1
        fi
    done
    echo "Done [elapsed:${timer}s]"
    return 0
}

function deploy() {
    local namespace=$1
    local pod_label=$2
    local resources_path=$3
    kubectl $(get_kubeconfig) apply -f "${resources_path}"
    if [[ $? -eq 0 ]]; then
        wait_for_pods ${namespace} ${pod_label} 20
        return $?
    fi
    echo "Failed to apply resources ${resources_path}"
    return -1
}


function delete() {
    local resources_path=$1
    kubectl $(get_kubeconfig) delete -f "${resources_path}" --ignore-not-found
    return $?
}
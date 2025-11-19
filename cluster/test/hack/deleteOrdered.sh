#!/bin/bash
set -eo pipefail

# Default variable values
KUBECTL=$1
CRD=$3
NAMESPACE=""

has_argument() {
    [[ ("$1" == *=* && -n ${1#*=}) || ( ! -z "$2" && "$2" != -*)  ]];
}

extract_argument() {
  echo "${2:-${1#*=}}"
}

# Function to handle options and arguments
handle_options() {
  while [ $# -gt 0 ]; do
    case $1 in
      -n | --namespace*)
        NAMESPACE=$(extract_argument $@)

        shift
        ;;
    esac
    shift
  done
}

# Main script execution
handle_options "$@"

retry_kubectl() {
  local max_attempts=10
  local delay=5
  local attempt=1
  local cmd="$1"
  local wait="$2"

  while [ $attempt -le $max_attempts ]; do
    echo "Kubectl attempt $attempt/$max_attempts for: $cmd"
    if eval "$cmd"; then
      echo "Kubectl operation successful on attempt $attempt"
      echo "Wait for deletion with: ${wait}"
      eval "${wait}"
      return 0
    else
      echo "Kubectl operation failed on attempt $attempt"
      if [ $attempt -lt $max_attempts ]; then
        echo "Retrying in ${delay}s..."
        sleep $delay
      fi
      ((attempt++))
    fi
  done
  echo "Kubectl operation failed after $max_attempts attempts"
  return 1
}

nsarg="--namespace ${NAMESPACE}"
if [ -z "${NAMESPACE}" ]
then
    nsarg=""
fi

retry_kubectl "${KUBECTL} delete ${CRD} --wait=false ${nsarg} --ignore-not-found" "${KUBECTL} wait ${nsarg} --for=delete ${CRD} --timeout 20m0s"

#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export TERM=dumb

OWD="${PWD}"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$( cd "${SCRIPT_DIR}/../../.." && pwd )"
CWD="${OWD}/_demo"
rm -rf "${CWD}"
mkdir -p "${CWD}"
cd "${CWD}"

logger -s "Creating Kind cluster"
kind create cluster --retain --config "${SCRIPT_DIR}/kind.conf.yaml" &
KIND_JOB="${!}"


logger -s "Getting Kube config"
until kind get kubeconfig > kube.config; do
    sleep 1
done

export KUBECONFIG="${PWD}/kube.config"

logger -s "Waiting for API server"
until kubectl get nodes; do
    sleep 1
done

logger -s "Copying Cluster CA"
docker cp kind-control-plane:/etc/kubernetes/pki/ca.key .
docker cp kind-control-plane:/etc/kubernetes/pki/ca.crt .

logger -s "Starting signer-ca"
${ROOT_DIR}/bin/manager \
   --ca-key-path=${PWD}/ca.key \
   --ca-cert-path=${PWD}/ca.crt \
   --signer-name=kubernetes.io/kube-apiserver-client-kubelet &
RUN_JOB="${!}"

logger -s "Waiting for kind create cluster to complete"
wait ${KIND_JOB}

logger -s "Stopping signer-ca"
kill "${RUN_JOB}"
wait

logger -s "Waiting for all nodes to be ready"
kubectl wait --for condition=Ready node --all

logger -s "Cluster ready"
kubectl get node

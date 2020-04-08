# K8S Bootstrap Configuration

This directory contains configuration for deploying `signer-ca` on the Kubernetes control-plane node in a Kind cluster.

It includes `kustomize` configuration and patches for deploying `signer-ca` with the cluster CA key and certificate.
And tolerations and node affinity to ensure that it deploys to the control-plane node.

# signer-ca


`signer-ca` is an operator for automatically signing an *approved* [CertificateSigningRequest](https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/#create-a-certificate-signing-request-object-to-send-to-the-kubernetes-api).

**NOTE: This operator is EXPERIMENTAL and requires Kubernetes >= 1.18.** It uses [Certificates API Enhancements](https://github.com/kubernetes/enhancements/blob/master/keps/sig-auth/20190607-certificates-api.md) which are only available in Kubernetes >= 1.18.

It watches `CertificateSigningRequest` (`CSR`) resources and if the `CSR` has a `.spec.signerName` that it recognizes,
and if the `CSR` has been approved,
it creates a signed certificate using a certificate-authority file that you supply as a command-line argument to the operator.
The signed certificate is configured using the encoded `CSR` in `.spec.request`.
The signed certificate is added to the `.status.certificate` of the `CSR` resource.

# Installation

`signer-ca` can be deployed using `kubectl apply -k config/default`.
See `config/e2e` for an example of how to make a `CA` file available to the operator, as a mounted secret.

# Build

You can build and deploy `signer-ca` using `make docker-build docker-push deploy-e2e DOCKER_PREFIX=gcr.io/<YOUR_PROJECT>/signer-ca/`.
See the `Makefile` for details.

# Demo

[![asciicast](https://asciinema.org/a/AxiLAeM8OUO6hXkRqGp4Z69N6.svg)](https://asciinema.org/a/AxiLAeM8OUO6hXkRqGp4Z69N6)

# Kubelet Signer Demo

This demonstrates how [signer-ca](https://github.com/cert-manager/signer-ca) can be used to sign the kubelet client certificate for the worker node in a multi-node [Kind](https://kind.sigs.k8s.io/) cluster.

Run the following command from the repository root:

```
make demo-kubelet-signer
```

## Notes

1. Disable the `csrcleaner` and `csrsigning` controllers in the Kubernetes controller-manager.
   See `kind.conf.yaml`.
2. Copy the cluster CA key and certificate out of the control-plane node, to a local directory.
3. `signer-ca` runs outside the cluster
4. `signer-ca` uses the copied CA key and cert
5. `signer-ca` is configured to sign Kubernetes CSRs.
   (signerName: kubernetes.io/kube-apiserver-client-kubelet)
6. Once the CSR has been signed (Issued) the worker node can connect to the API server and `kubeadm` via `kind` can complete the cluster.


[![asciicast](https://asciinema.org/a/j8i7Ms19oBeox4VhNETaeEPWS.svg)](https://asciinema.org/a/j8i7Ms19oBeox4VhNETaeEPWS)

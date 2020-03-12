module github.com/cert-manager/signer-ca

go 1.14

// Pin k8s.io/* dependencies to kubernetes-1.17.0 to match controller-runtime v0.5.0
replace (
	k8s.io/api => k8s.io/api v0.18.0-beta.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/apiserver => k8s.io/apiserver v0.18.0-beta.2
	k8s.io/client-go => k8s.io/client-go v0.17.3
)

require (
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2 // indirect
	k8s.io/api v0.18.0-beta.2
	k8s.io/apimachinery v0.18.0-beta.2
	k8s.io/apiserver v0.17.2
	k8s.io/client-go v0.18.0-beta.2
	sigs.k8s.io/controller-runtime v0.5.0
)

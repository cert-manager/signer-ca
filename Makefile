MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
.SUFFIXES:

# The semver version number which will be used as the Docker image tag
# Defaults to the output of git describe.
VERSION ?= $(shell git describe --tags --dirty --always)

# Docker image name parameters
DOCKER_PREFIX ?= quay.io/cert-manager/signer-ca-
DOCKER_TAG ?= ${VERSION}
DOCKER_IMAGE ?= ${DOCKER_PREFIX}controller:${DOCKER_TAG}

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

ARGS ?=

BIN := ${CURDIR}/bin
export PATH := ${BIN}:${PATH}

all: manager

# Run tests
test:
	go test ./... -coverprofile cover.out

# Build manager binary
manager:
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run:
	go run ./main.go ${ARGS}

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${DOCKER_IMAGE}
	kustomize build config/default | kubectl apply -f -

E2E_PKI = config/e2e
E2E_CA_CERT = ${E2E_PKI}/tls.crt
E2E_CA_KEY = ${E2E_PKI}/tls.key
E2E_CA = ${E2E_CA_KEY} ${E2E_CA_CERT}

${E2E_CA}:
	mkdir -p ${E2E_PKI}
	cd ${E2E_PKI} && cfssl gencert -initca ca-csr.json | cfssljson -bare tls
	mv ${E2E_PKI}/tls.{pem,crt}
	mv ${E2E_PKI}/tls{-key.pem,.key}

deploy-e2e: ${E2E_CA}
	cd config/e2e && kustomize edit set image controller=${DOCKER_IMAGE}
	kustomize build config/e2e | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build:
	docker build . -t ${DOCKER_IMAGE}

# Push the docker image
docker-push:
	docker push ${DOCKER_IMAGE}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

.PHONY: demo-kubelet-signer
demo-kubelet-signer: manager
	docs/demos/kubelet-signer/kubelet-signer-demo.sh

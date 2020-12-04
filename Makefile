# note: call scripts from /scripts

.PHONY: default verify build builder-image binary-image test stop clean-images clean push apply deploy helm-template helm-install

BUILDER ?= ingressmonitorcontroller-builder
BINARY ?= IngressMonitorController
DOCKER_IMAGE ?= stakater/ingressmonitorcontroller

# GOLANGCI_LINT env
GOLANGCI_LINT = _output/tools/golangci-lint
GOLANGCI_LINT_CACHE = $(PWD)/_output/golangci-lint-cache
GOLANGCI_LINT_VERSION = v1.24

# Default value "dev"
DOCKER_TAG ?= dev
REPOSITORY = ${DOCKER_IMAGE}:${DOCKER_TAG}

VERSION ?= 0.0.1
BUILD=

GOCMD = go
GOFLAGS ?= $(GOFLAGS:)
GOMAINPACKAGE=./cmd/manager

LDFLAGS =

HELMPATH= deployments/kubernetes/chart/ingressmonitorcontroller
HELMVALUES = $(HELMPATH)/values.yaml
HELMNAME = IMC

default: build test

install:
	"$(GOCMD)" mod download

build:
	"$(GOCMD)" build ${GOFLAGS} ${LDFLAGS} -o "${BINARY}" $(GOMAINPACKAGE)

verify-fmt:
	./hack/verify-gofmt.sh

$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(dir $@) v1.24.0

verify-golangci-lint: $(GOLANGCI_LINT)
	GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) $(GOLANGCI_LINT) run --timeout=300s ./cmd/... ./pkg/...

verify: verify-fmt verify-golangci-lint

builder-image:
	@docker build --network host -t "${REPOSITORY}" -f build/Dockerfile .

binary-image: builder-image

test:
	GOFLAGS="-count=1" "$(GOCMD)" test -v ./...

run-local:
	./hack/run-local.sh

update-resources-ci: update-operator-image
	CSV_VERSION=$(VERSION) ./hack/update-resources.sh

update-operator-image:
	sed -i "s|image: stakater/ingressmonitorcontroller:v.*|image: stakater/ingressmonitorcontroller:v$(VERSION)|" deploy/operator.yaml

stop:
	@docker stop "${BINARY}"

clean-images: stop
	@docker rmi "${BUILDER}" "${BINARY}"

clean:
	"$(GOCMD)" clean -i

push: ## push the latest Docker image to DockerHub
	docker push $(REPOSITORY)

apply:
	kubectl apply -f deployments/manifests/

deploy: binary-image push apply

helm-template:
	helm template $(HELMPATH) --values $(HELMVALUES) --name $(HELMNAME)

helm-install:
	helm install $(HELMPATH) --values $(HELMVALUES) --name $(HELMNAME)

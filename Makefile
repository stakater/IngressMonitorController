# note: call scripts from /scripts

.PHONY: default build builder-image binary-image test stop clean-images clean push apply deploy helm-template helm-install

BUILDER ?= ingressmonitorcontroller-builder
BINARY ?= IngressMonitorController
DOCKER_IMAGE ?= stakater/ingressmonitorcontroller

# Default value "dev"
DOCKER_TAG ?= dev
REPOSITORY = ${DOCKER_IMAGE}:${DOCKER_TAG}

VERSION=$(shell cat .version)
BUILD=

GOCMD = go
GLIDECMD = glide
GOFLAGS ?= $(GOFLAGS:)
LDFLAGS =

HELMPATH= deployments/kubernetes/chart/ingressmonitorcontroller
HELMVALUES = $(HELMPATH)/values.yaml
HELMNAME = IMC

default: build test

install:
	"$(GLIDECMD)" update --strip-vendor

build:
	"$(GOCMD)" build ${GOFLAGS} ${LDFLAGS} -o "${BINARY}"

builder-image:
	@docker build --network host -t "${BUILDER}" -f build/package/Dockerfile.build .

binary-image: builder-image
	@docker run --network host --rm "${BUILDER}" | docker build --network host -t "${REPOSITORY}" -f Dockerfile.run -

test:
	"$(GOCMD)" test -v ./...

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

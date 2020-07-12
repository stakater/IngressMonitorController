FROM registry.svc.ci.openshift.org/openshift/release:golang-1.10 AS builder
COPY . /go/src/github.com/stakater/IngressMonitorController
WORKDIR /go/src/github.com/stakater/IngressMonitorController
RUN make gobuild

FROM registry.svc.ci.openshift.org/openshift/origin-v4.0:base
ENV OPERATOR_BIN=deadmanssnitch-operator

WORKDIR /root/
COPY --from=builder /go/src/github.com/stakater/IngressMonitorController/build/_output/bin/${OPERATOR_BIN} /usr/local/bin/${OPERATOR_BIN}
MAINTAINER "Stakater Team"

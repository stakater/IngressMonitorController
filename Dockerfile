FROM stakater/go-glide:1.9.3
LABEL author="stakater"

ARG SRC_DIR="ingress-monitor-controller"
ENV SRC_DIR=${SRC_DIR}

ADD bootstrap.sh /bootstrap.sh

ADD ./src /go/${SRC_DIR}

CMD [ "/bin/sh", "-c", "/bootstrap.sh" ]
FROM stakater/go-glide:1.9.3
LABEL author="stakater"

ARG SRC_DIR="ingress-monitor-controller"
ENV SRC_DIR=${SRC_DIR}

ADD ./src /go/${SRC_DIR}
RUN cd ${SRC_DIR} && \
    glide update && \
    cp -r ./vendor/* /go/src/ && \
    go build -o ./out/main

CMD [ "/bin/sh", "-c", "${SRC_DIR}/out/main" ]
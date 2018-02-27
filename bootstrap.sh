#!/usr/bin/env sh

cd ${SRC_DIR}

glide update

cp -r ./vendor/* /go/src/

go test && \

go build -o ./out/main

./main
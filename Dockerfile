FROM stakater/base-alpine:3.7
LABEL author="stakater"

COPY ./out/ingressmonitorcontroller /

ENTRYPOINT [ "/ingressmonitorcontroller" ]
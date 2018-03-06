FROM scratch
LABEL author="stakater"

COPY ./out/ingressmonitorcontroller /

ENTRYPOINT [ "/ingressmonitorcontroller" ]
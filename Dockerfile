FROM scratch
LABEL author="stakater"

COPY /out/ingressmonitorcontroller /ingressmonitorcontroller

ENTRYPOINT [ "/ingressmonitorcontroller" ]
FROM alpine:3.11

COPY kube-eventer /kube-eventer

ENTRYPOINT ["./kube-eventer"]

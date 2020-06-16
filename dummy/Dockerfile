ARG IMAGE=golang:latest
FROM $IMAGE

RUN go get github.com/go-delve/delve/cmd/dlv

ENTRYPOINT ["/bin/sh", "-c", "while true; do echo dummy!; sleep 3600; done"]
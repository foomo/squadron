ARG BASE_IMAGE=golang
ARG BASE_IMAGE_TAG=alpine
FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG} as builder

ENV CGO_ENABLED=0

COPY / /src

WORKDIR /src

RUN go build -ldflags "-w -s" -trimpath -o /go/bin/service .

FROM alpine:latest

COPY --from=builder /go/bin/service /usr/local/bin/service

ENTRYPOINT ["/usr/local/bin/service"]

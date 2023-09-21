FROM golang:alpine as builder

ENV CGO_ENABLED=0

COPY / /src

WORKDIR /src

RUN go build -ldflags "-w -s" -trimpath -o /go/bin/service .

FROM alpine:latest as development

COPY --from=builder /go/bin/service /usr/local/bin/service

ENTRYPOINT ["/usr/local/bin/service"]

FROM golang:alpine AS builder
COPY test/ .
RUN go get -d -v
RUN go build -o /go/bin/hello

FROM scratch
COPY --from=builder /go/bin/hello /go/bin/hello
ENTRYPOINT ["/go/bin/hello"]
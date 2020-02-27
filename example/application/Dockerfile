FROM golang:alpine AS builder

COPY . .

RUN go build -o /hello .

FROM alpine:latest

COPY --from=builder /hello /hello

ENTRYPOINT ["/hello"]
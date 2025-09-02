FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the Go application
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -tags=safe -o /usr/bin/squadron ./cmd/main.go

# Stage 3: Final image
FROM alpine:latest

RUN adduser -D -u 12345 -g 12345 squadron
COPY --from=builder /usr/bin/squadron /usr/bin/squadron

USER 12345
WORKDIR /home/squadron

ENTRYPOINT ["squadron"]

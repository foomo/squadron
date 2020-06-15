if [ "$1" = "golang:alpine" ]; then
    echo "installing delve"
    apk add --no-cache git
    go get github.com/go-delve/delve/cmd/dlv
fi
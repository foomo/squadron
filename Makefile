pack-data:
	go-bindata -o bindata/bindata.go -pkg bindata example/... dummy/...

build: pack-data
	go build -o /usr/local/bin/configurd cmd/main.go

test:
	go test ./...
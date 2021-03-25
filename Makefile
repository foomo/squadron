pack-example:
	go-bindata -prefix=example -o exampledata/exampledata.go -pkg exampledata example/...

install: pack-example 
	go build -o /usr/local/bin/squadron cmd/main.go

build: pack-example
	go build -o bin/squadron cmd/main.go

test:
	go test ./...

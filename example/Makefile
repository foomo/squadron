build:
	go run ../cmd/main.go build hello-service -t master

install:
	go run ../cmd/main.go install local hello-group -t master --build

uninstall:
	go run ../cmd/main.go uninstall local -g hello-group

install-olaf:
	go run ../cmd/main.go install olaf-group -n local -t latest --build -v

uninstall-olaf:
	go run ../cmd/main.go uninstall olaf-group -n local

VERSION=0.0.3
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"
all: mackerel-plugin-linux-netdev

.PHONY: mackerel-plugin-linux-netdev linux check lint

mackerel-plugin-linux-netdev: *.go
	go build $(LDFLAGS) -o mackerel-plugin-linux-netdev

linux: *.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o mackerel-plugin-linux-netdev

check:
	go test -v ./...

lint:
	golangci-lint run --timeout 5m ./...
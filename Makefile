export GO111MODULE=on

.PHONY: build static

build: static
	go build -o bin/kube-daemonset-proxy

install-tools:
	GO111MODULE=off go get github.com/rakyll/statik

static:
	statik -src=static

DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
DIST_DIR=dist
BINARY=ces-confd

.PHONY: default setup env build exec run daemon sample vet clean

default: build

setup:
	@go get gopkg.in/yaml.v2
	@go get github.com/codegangsta/cli
	@go get github.com/coreos/etcd/client

build: setup
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -ldflags "-linkmode external -extldflags -static" -o $(DIST_DIR)/$(BINARY)

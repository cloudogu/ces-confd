DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

DIST_DIR=dist
BINARY=ces-confd

# These are the values we want to pass for Version and BuildTime
VERSION=0.1.5

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-linkmode external -extldflags -static -X main.Version=${VERSION}"

$(BINARY):
	@echo "build ..."
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo ${LDFLAGS} -o $(DIST_DIR)/$(BINARY)
	echo "... binary can be found at $(DIST_DIR)/$(BINARY)"

.PHONY: package
package: $(BINARY)
	@echo "create package ..."
	cd $(DIST_DIR) && tar cvfz $(BINARY)-v$(VERSION).tar.gz $(BINARY)

.PHONY: clean
clean:
	@echo "clean ..."
	cd ${DIR}; rm -rf ${DIST_DIR}

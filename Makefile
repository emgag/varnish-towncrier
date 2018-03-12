.PHONY: build install test vet lint fmt run clean docker
OUT := varnish-towncrier
PKG := github.com/emgag/varnish-towncrier
VERSION := $(shell git describe --always --dirty --tags --long)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)

all: build

build:
	CGO_ENABLED=0 GOOS=linux go build -a -v -o ${OUT} -ldflags="-X ${PKG}/internal/lib.Version=${VERSION}" ${PKG}

install:
	CGO_ENABLED=0 GOOS=linux go install -a -v -o ${OUT} -ldflags="-X ${PKG}/internal/lib.Version=${VERSION}" ${PKG}

test:
	@go test -v ${PKG_LIST}

vet:
	@go vet ${PKG_LIST}

lint:
	@for file in ${GO_FILES} ;  do \
		golint $$file ; \
	done

fmt:
	@gofmt -l -w -s ${GO_FILES}

run: build
	./${OUT} listen

clean:
	-@rm ${OUT}

docker: build
	docker build \
		-t emgag/varnish-towncrier:${VERSION} \
		-t emgag/varnish-towncrier:latest\
		.
	docker push emgag/varnish-towncrier:${VERSION}
	docker push emgag/varnish-towncrier:latest



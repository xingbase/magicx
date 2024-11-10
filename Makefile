ifeq ($(OS), Windows_NT)
	VERSION := $(shell git describe --exact-match --tags 2>nil)
else
	VERSION := $(shell git describe --exact-match --tags 2>/dev/null)
endif

COMMIT ?= $(shell git rev-parse --short=8 HEAD)

unexport LDFLAGS
ifdef VERSION
	TMP_BUILD_VERSION = -X main.version=$(VERSION)
endif
LDFLAGS=-ldflags "-s -X main.commit=${COMMIT} ${TMP_BUILD_VERSION}"
unexport TMP_BUILD_VERSION

BINARY=magicx

build:
	# CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -installsuffix cgo -o bin/${BINARY} ./cmd/magicx/main.go
	go build -o bin/${BINARY} ${LDFLAGS} ./cmd/magicx/*.go
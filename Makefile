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

build: build-win build-mac

build-win:
	CGO_ENABLED=1 GODEBUG=cgocheck=0 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -v -o bin/${BINARY}.exe ./cmd/gui/main.go

build-mac:
	go build -o bin/${BINARY} ${LDFLAGS} ./cmd/gui/main.go
.PHONY: build depend install test lint clean vet

PROG:=autoscale
BUILD_DIR:=.build
DIST_DIR:=.dist
TARGET:=$(BUILD_DIR)/$(PROG)

VERSION:=0.1
BUILD:=$(shell git rev-parse HEAD)
GIT_TAG:=$(shell git describe --exact-match HEAD 2>/dev/null)
UNCOMMITED_CHANGES:=$(shell git diff-index --shortstat HEAD 2>/dev/null)

ifeq (v$(VERSION), $(GIT_TAG))
BUILD_TYPE:=RELEASE
else
BUILD_TYPE:=SNAPSHOT
endif

ifneq ($(strip $(UNCOMMITED_CHANGES)),)
BUILD_TYPE:=DEV
BUILD_DATE:=$(shell date +%FT%T%z)
endif

GOPATH ?= $(HOME)/go

default: build

build: vet
	go build -v -buildmode=exe -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildType=$(BUILD_TYPE) -X main.Build=$(BUILD) -X main.BuildDate=$(BUILD_DATE)" -o $(TARGET)

clean:
	go clean -i ./... && \
if [ -d $(BUILD_DIR) ] ; then rm -rf $(BUILD_DIR) ; fi && \
if [ -d $(DIST_DIR) ] ; then rm -rf $(DIST_DIR) ; fi

depend:
	go get -u -ldflags "-s -w" github.com/streadway/amqp
	go get -u -ldflags "-s -w" github.com/mattn/go-sqlite3
	go get -u -ldflags "-s -w" k8s.io/client-go/...
	if [ -d $(GOPATH)/src/k8s.io/kubernetes ] ; then rm -rf $(GOPATH)/src/k8s.io/kubernetes ; fi && git clone --depth 1 -b v1.7.8 --single-branch -q https://github.com/kubernetes/kubernetes.git $(GOPATH)/src/k8s.io/kubernetes


install:
	go install $(TARGET)

lint:
	golint ./...

test:
	go test -v ./...

vet:
	go tool vet -all -shadow *.go

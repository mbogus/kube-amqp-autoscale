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
	go get -u -ldflags "-s -w" github.com/googleapis/gnostic/openapiv2
	go get -u -ldflags "-s -w" github.com/gogo/protobuf/proto
	go get -u -ldflags "-s -w" github.com/gogo/protobuf/sortkeys
	go get -u -ldflags "-s -w" github.com/davecgh/go-spew/spew
	go get -u -ldflags "-s -w" github.com/google/gofuzz
	go get -u -ldflags "-s -w" github.com/json-iterator/go
	go get -u -ldflags "-s -w" github.com/modern-go/reflect2
	go get -u -ldflags "-s -w" golang.org/x/crypto/ssh/terminal
	go get -u -ldflags "-s -w" golang.org/x/net/http2
	go get -u -ldflags "-s -w" golang.org/x/oauth2
	go get -u -ldflags "-s -w" golang.org/x/time/rate
	go get -u -ldflags "-s -w" gopkg.in/inf.v0
	go get -u -ldflags "-s -w" github.com/prometheus/client_golang/prometheus
	go get -u -ldflags "-s -w" github.com/prometheus/client_golang/prometheus/promhttp
	if [ -d $(GOPATH)/src/sigs.k8s.io/yaml ] ; then rm -rf $(GOPATH)/src/sigs.k8s.io/yaml ; fi && git clone --depth 1 -b v1.1.0 --single-branch -q https://github.com/kubernetes-sigs/yaml $(GOPATH)/src/sigs.k8s.io/yaml
	if [ -d $(GOPATH)/src/k8s.io/klog ] ; then rm -rf $(GOPATH)/src/k8s.io/klog ; fi && git clone --depth 1 -b v0.4.0 --single-branch -q https://github.com/kubernetes/klog $(GOPATH)/src/k8s.io/klog
	if [ -d $(GOPATH)/src/k8s.io/client-go ] ; then rm -rf $(GOPATH)/src/k8s.io/client-go ; fi && git clone --depth 1 -b v0.17.3 --single-branch -q https://github.com/kubernetes/client-go $(GOPATH)/src/k8s.io/client-go
	if [ -d $(GOPATH)/src/k8s.io/utils ] ; then rm -rf $(GOPATH)/src/k8s.io/utils ; fi && git clone --depth 1 -q https://github.com/kubernetes/utils $(GOPATH)/src/k8s.io/utils
	if [ -d $(GOPATH)/src/k8s.io/api ] ; then rm -rf $(GOPATH)/src/k8s.io/api ; fi && git clone --depth 1 -b v0.17.3 --single-branch -q https://github.com/kubernetes/api $(GOPATH)/src/k8s.io/api
	if [ -d $(GOPATH)/src/k8s.io/apimachinery ] ; then rm -rf $(GOPATH)/src/k8s.io/apimachinery ; fi && git clone --depth 1 -b v0.17.3 --single-branch -q https://github.com/kubernetes/apimachinery $(GOPATH)/src/k8s.io/apimachinery
	if [ -d $(GOPATH)/src/k8s.io/kubernetes ] ; then rm -rf $(GOPATH)/src/k8s.io/kubernetes ; fi && git clone --depth 1 -b v1.17.3 --single-branch -q https://github.com/kubernetes/kubernetes $(GOPATH)/src/k8s.io/kubernetes


install:
	go install $(TARGET)

lint:
	golint ./...

test:
	go test -v ./...

vet:
	go vet -all *.go

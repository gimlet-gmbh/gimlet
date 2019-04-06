# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

CLI_BINARY=gmbh
CORE_BINARY=gmbhCore
PROCM_BINARY=gmbhProcm

all: cli core procm

cli: build-cli install-cli
core: build-core install-core
procm: build-procm install-procm

build-cli:
	$(GOBUILD) -o ./bin/$(CLI_BINARY) ./cmd/gmbh/*.go
build-core:
	$(GOBUILD) -o ./bin/$(CORE_BINARY) ./cmd/gmbhCore/*.go
build-procm:
	$(GOBUILD) -o ./bin/$(PROCM_BINARY) ./cmd/gmbhProcm/*.go

install-cli:
	cp bin/$(CLI_BINARY) "${GOPATH}/bin"
install-core:
	cp bin/$(CORE_BINARY) "${GOPATH}/bin"
install-procm:
	cp bin/$(PROCM_BINARY) "${GOPATH}/bin"


deps:
	$(GOGET) -u github.com/golang/protobuf/proto
	$(GOGET) -u github.com/golang/protobuf/protoc-gen-go 
	$(GOGET) -u google.golang.org/grpc
	$(GOGET) -u github.com/BurntSushi/toml
	$(GOGET) -u github.com/fatih/color
	$(GOGET) -u github.com/rs/xid
	
clean: 
	rm -f ./bin/*

.PONY:

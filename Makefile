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
	cp bin/$(CLI_BINARY) /usr/local/bin/
install-core:
	cp bin/$(CORE_BINARY) /usr/local/bin/
install-procm:
	cp bin/$(PROCM_BINARY) /usr/local/bin/


deps:
	$(GOGET) -u github.com/golang/protobuf/proto
	$(GOGET) -u github.com/golang/protobuf/protoc-gen-go 
	$(GOGET) -u google.golang.org/grpc
	$(GOGET) -u github.com/fatih/color
	$(GOGET) -u gopkg.in/yaml.v2
	
clean: 
	rm -f ./bin/*

.PONY:

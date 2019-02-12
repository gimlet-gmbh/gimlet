# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

CLI_BINARY=gmbh
CORE_BINARY=gmbhCore
CTRL_BINARY=gmbhCtrl

all: build install


build:
	$(GOBUILD) -o ./bin/$(CORE_BINARY) ./cmd/gmbhCore/*.go
	$(GOBUILD) -o ./bin/$(CLI_BINARY) ./cmd/gmbh/*.go
	$(GOBUILD) -o ./bin/$(CTRL_BINARY) ./cmd/gmbhCtrl/*.go

install:
	cp bin/$(CORE_BINARY) /usr/local/bin/
	cp bin/$(CLI_BINARY) /usr/local/bin/
	cp bin/$(CTRL_BINARY) /usr/local/bin/

# test: 
# 	$(GOTEST) -v ./...

deps:
	$(GOGET) github.com/fatih/color
	$(GOGET) google.golang.org/grpc
	$(GOGET) gopkg.in/yaml.v2
	$(GOGET) github.com/rs/xid
	$(GOGET) github.com/golang/protobuf/proto
	$(GOGET) github.com/golang/protobuf/protoc-gen-go 
	$(GOGET) github.com/golang/protobuf/{proto,protoc-gen-go}
	
clean: 
	rm -f ./bin/$(BINARY_NAME)


.PONY:
	osx

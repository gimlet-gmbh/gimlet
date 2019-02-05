# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

CLI_BINARY=gmbh
CORE_BINARY=gmbhCore
CTRL_BINARY=gmbhCtrl
CONTAINER_BINARY=gmbhContainer

all: build-core build-cli build-ctrl build-container install-core install-cli install-ctrl install-container

core: build-core install-core
cli: build-cli install-cli
ctrl: build-ctrl install-ctrl
container: build-container install-container

build-core:
	$(GOBUILD) -o ./bin/$(CORE_BINARY) ./cmd/gmbhCore/*.go
build-cli:
	$(GOBUILD) -o ./bin/$(CLI_BINARY) ./cmd/gmbh/*.go
build-ctrl:
	$(GOBUILD) -o ./bin/$(CTRL_BINARY) ./cmd/gmbhCtrl/*.go
build-container:
	$(GOBUILD) -o ./bin/$(CONTAINER_BINARY) ./cmd/gmbhContainer/*.go

install-core:
	cp bin/$(CORE_BINARY) /usr/local/bin/
install-cli:
	cp bin/$(CLI_BINARY) /usr/local/bin/
install-ctrl:
	cp bin/$(CTRL_BINARY) /usr/local/bin/
install-container:
	cp bin/$(CONTAINER_BINARY) /usr/local/bin/

# test: 
# 	$(GOTEST) -v ./...

deps:
	$(GOGET) github.com/fatih/color
	$(GOGET) google.golang.org/grpc
	$(GOGET) gopkg.in/yaml.v2
	$(GOGET) github.com/rs/xid
	$(GOGET) github.com/golang/protobuf/proto
	$(GOGET) github.com/golang/protobuf/protoc-gen-go 

clean: 
	rm -f ./bin/$(BINARY_NAME)


.PONY:
	osx

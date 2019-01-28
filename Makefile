# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

CLI_BINARY=gmbh
CORE_BINARY=gmbhCore
CTRL_BINARY=gmbhCtrl

all: build-core build-cli build-ctrl

install: all install-core install-cli install-ctrl

build-core:
	$(GOBUILD) -o ./bin/$(CORE_BINARY) ./cmd/gmbhCore/*.go

build-cli:
	$(GOBUILD) -o ./bin/$(CLI_BINARY) ./cmd/gmbh/*.go

build-ctrl:
	$(GOBUILD) -o ./bin/$(CTRL_BINARY) ./cmd/gmbhCtrl/*.go

install-core:
	cp bin/$(CORE_BINARY) /usr/local/bin/

install-cli:
	cp bin/$(CLI_BINARY) /usr/local/bin/

install-ctrl:
	cp bin/$(CTRL_BINARY) /usr/local/bin/

# test: 
# 	$(GOTEST) -v ./...

deps:
	$(GOGET) github.com/fatih/color
	$(GOGET) google.golang.org/grpc
	$(GOGET) gopkg.in/yaml.v2
	$(GOGET) github.com/rs/xid
	$(GOGET) github.com/golang/protobuf/proto

clean: 
	rm -f ./bin/$(BINARY_NAME)
run:
	$(GOBUILD) -o ./bin/$(BINARY_NAME)  ./src
	./bin/$(BINARY_NAME)

.PONY:
	install-core install-cli install-ctrl

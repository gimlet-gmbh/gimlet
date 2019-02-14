# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

CLI_BINARY=gmbh
CORE_BINARY=gmbhCore
CTRL_BINARY=gmbhCtrl
PM_BINARY=gmbhPM
CONTAINER_BINARY=gmbhContainer
NC_BINARY=gmbhNC

all: core cli ctrl pm nc

core: build-core install-core
cli: build-cli install-cli
ctrl: build-ctrl install-ctrl
pm: build-pm install-pm
nc: build-nc install-nc

build-core:
	$(GOBUILD) -o ./bin/$(CORE_BINARY) ./cmd/gmbhCore/*.go
build-cli:
	$(GOBUILD) -o ./bin/$(CLI_BINARY) ./cmd/gmbh/*.go
build-ctrl:
	$(GOBUILD) -o ./bin/$(CTRL_BINARY) ./cmd/gmbhCtrl/*.go
build-pm:
	$(GOBUILD) -o ./bin/$(PM_BINARY) ./cmd/gmbh_pm/*.go
build-nc:
	$(GOBUILD) -o ./bin/$(NC_BINARY) ./cmd/gmbh_nc/*.go

install-core:
	cp bin/$(CORE_BINARY) /usr/local/bin/
install-cli:
	cp bin/$(CLI_BINARY) /usr/local/bin/
install-ctrl:
	cp bin/$(CTRL_BINARY) /usr/local/bin/
install-pm:
	cp bin/$(PM_BINARY) /usr/local/bin/
install-nc:
	cp bin/$(NC_BINARY) /usr/local/bin/

deps:
	$(GOGET) github.com/fatih/color
	$(GOGET) google.golang.org/grpc
	$(GOGET) gopkg.in/yaml.v2
	$(GOGET) github.com/golang/protobuf/proto
	$(GOGET) github.com/golang/protobuf/protoc-gen-go 
	$(GOGET) github.com/golang/protobuf/{proto,protoc-gen-go}
	
clean: 
	rm -f ./bin/$(BINARY_NAME)

.PONY:

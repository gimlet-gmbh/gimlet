#!/bin/bash


# if [ $GOPATH == ""]; then
#     echo "GOPATH is not set. Will attempt to use GOPATH=$HOME/go"
#     export GOPATH=$HOME/go
# fi

if [ $GOPATH == ""]; then
    echo "GOPATH must be set before installing gmbh packages."
    echo "To set to the default GOPATH add \`export GOPATH=\$HOME/go\` to your bash profile"
    exit
fi

./link_packages.sh
cd ../

## Install Deps
echo "installing deps"
go get github.com/golang/protobuf/proto
go get github.com/golang/protobuf/protoc-gen-go 
go get google.golang.org/grpc
go get github.com/fatih/color
go get gopkg.in/yaml.v2

## Build Binaries
echo "building gmbh"
go build -v -o ./bin/gmbh ./cmd/gmbh/*.go
echo "building gmbhCore"
go build -v -o ./bin/gmbhCore ./cmd/gmbhCore/*.go
echo "building gmbhProcm"
go build -v -o ./bin/gmbhProcm ./cmd/gmbhProcm/*.go

## Copy to bin
echo "copying files to /usr/local/bin"
cp ./bin/gmbh* /usr/local/bin

echo "done"
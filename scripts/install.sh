#!/bin/bash


# if [ $GOPATH == ""]; then
#     echo "GOPATH is not set. Will attempt to use GOPATH=$HOME/go"
#     export GOPATH=$HOME/go
# fi
IFS=''
if [ $GOPATH == ""]; then
    echo "GOPATH must be set before installing gmbh packages."
    echo "\"\$GOPATH/bin\" must also be set to run gmbh from the command line."
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
go get github.com/rs/xid

## Build Binaries
echo "building gmbh"
go build -v -o ./bin/gmbh ./cmd/gmbh/*.go
echo "building gmbhCore"
go build -v -o ./bin/gmbhCore ./cmd/gmbhCore/*.go
echo "building gmbhProcm"
go build -v -o ./bin/gmbhProcm ./cmd/gmbhProcm/*.go

## Copy to bin
echo "copying files to $GOPATH/bin"
cp ./bin/gmbh* $GOPATH/bin

echo "done"
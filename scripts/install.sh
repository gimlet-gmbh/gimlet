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
make deps


## Experiment for protoc
# cd ../internal/cabal
# rm cabal.pb.go
# ./build.sh
# cd ../../


make

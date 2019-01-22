#!/bin/bash

echo "installing gmbh cli at: " $GOPATH"/bin"

cd ../
make build-cli

cp ./bin/gmbh $(echo $GOPATH)/bin/
#!/bin/bash

IFS=''

if [ $GOPATH == ""]; then
    echo "GOPATH must be set before installing gmbh packages."
    exit
fi

echo "uninstalling old versions of gmbh-micro"
rm -rf $GOPATH/src/github.com/gmbh-micro
rm -rf $GOPATH/src/github.com/gimlet-gmbh

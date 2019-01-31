#!/bin/bash

if [ $GOPATH != "" ]; then
    echo "uninstalling old versions of gmbh-micro"
    rm -rf $GOPATH/src/github.com/gmbh-micro
    rm -rf $GOPATH/src/github.com/gimlet-gmbh
else 
    echo "Rerun after setting your go path"
fi
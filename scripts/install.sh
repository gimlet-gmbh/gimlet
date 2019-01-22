#!/bin/bash

if [ $GOPATH != "" ]; then
    ./link_packages.sh
    cd ../
    make deps
    make install
else 
    echo "Rerun after setting your go path"
fi
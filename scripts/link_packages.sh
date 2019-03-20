#!/bin/bash

# Creates a symbolic link to all folders in ./internal to the correct destination
# so that the Go Toolchain can find them without changing your GOPATH.
IFS=''
if [ $GOPATH == ""]; then
    echo "GOPATH must be set before installing gmbh packages."
    exit
fi

# Link packages needed to build gmbh server and Go package
GMBH_PATH=$GOPATH"/src/github.com/gmbh-micro"
mkdir -p $GMBH_PATH

echo "linking gmbh at $GMBH_PATH"

# Link the internal packages for building the core
for dir in ../internal/*/
do
    dir=${dir%*/}
    PKG_PATH=$GMBH_PATH"/"${dir##*/}
    if [ -d $PKG_PATH ]; then
        echo ${dir##*/}" is already linked"
    else
        echo "Linking: "${dir##*/}
        pdir=$(pwd)/$dir
        ln -s ${pdir} $GMBH_PATH
    fi
done


# Link the go client package
if [ -d $GMBH_PATH/gmbh ]; then
    echo "gmbh is already linked"
else
    ln -s $PWD"/../pkg/gmbh" $GMBH_PATH"/gmbh"
    echo "linking gmbh to go path"
fi

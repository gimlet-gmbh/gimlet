#!/bin/bash

# Creates a symbolic link to all folders in ./internal to the correct destination
# so that the Go Toolchain can find them without changing your GOPATH.

if [ $GOPATH != "" ]; then
    gPATH=$GOPATH"/src/github.com/gimlet-gmbh/gimlet/"
    PKGPATH=$GOPATH"/src/github.com/gimlet-gmbh"
    echo "Linking with Gopath="$GOPATH
    echo $gPATH

    mkdir -p $gPATH
    for dir in ../internal/*/
    do  
        dir=${dir%*/}
        pkg=$gPATH${dir##*/}
        if [ -d $pkg ]; then
            echo ${dir##*/}" is already linked"
        else
            echo "Linking: "${dir##*/}
            pdir=$(pwd)/$dir
            ln -s ${pdir} $gPATH
        fi
    done

    # Link the gmbh client package
    if [ -d $PKGPATH/gmbh ]; then
        echo "gmbh is already linked"
    else 
        ln -s ../pkg/gmbh-go $PKGPath/gmbh
        echo "linking gmbh to go path"
    fi
    

else 
    echo "Please set the GOPATH environment variable and rerun."
fi

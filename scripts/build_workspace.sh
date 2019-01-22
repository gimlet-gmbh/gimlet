#!/bin/bash

cd ../

mkdir -p bin
mkdir -p cmd
mkdir -p pkg
mkdir -p services

cd cmd
git clone https://github.com/gimlet-gmbh/gimlet-core.git
git clone https://github.com/gimlet-gmbh/cli.git

cd ../

git clone https://github.com/gimlet-gmbh/gimlet.git
mv gimlet/ internal/

cd ../
cd pkg

git clone https://github.com/gimlet-gmbh/gmbh-go.git
git clone https://github.com/gimlet-gmbh/gmbh-python.git
git clone https://github.com/gimlet-gmbh/gmbh-node.git

cd ../
cd services

git clone https://github.com/gimlet-gmbh/gmbh-webserver.git
git clone https://github.com/gimlet-gmbh/gmbh-demo.git

./link_packages.sh
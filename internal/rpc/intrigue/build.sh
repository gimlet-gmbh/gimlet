#!/bin/bash

# Go
protoc --go_out=plugins=grpc:. *.proto

cd ../
# Python
python -m grpc_tools.protoc -I./intrigue --python_out=. --grpc_python_out=. ./intrigue/intrigue.proto
mv *.py ./intrigue/

# Node Version
protoc --js_out=import_style=commonjs,binary:./ intrigue/intrigue.proto

grpc_tools_node_protoc --js_out=import_style=commonjs,binary:./ --grpc_out=./ --plugin=protoc-gen-grpc=`which grpc_tools_node_protoc_plugin` ./intrigue/intrigue.proto
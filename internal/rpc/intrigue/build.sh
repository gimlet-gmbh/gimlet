#!/bin/bash


# Go
protoc --go_out=plugins=grpc:. *.proto

cd ../
# Python
python -m grpc_tools.protoc -I./intrigue --python_out=. --grpc_python_out=. ./intrigue/intrigue.proto
mv *.py ./intrigue/

# Node Version
protoc --js_out=import_style=commonjs,binary:./ intrigue/intrigue.proto

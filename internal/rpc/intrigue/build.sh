#!/bin/bash

protoc --go_out=plugins=grpc:. *.proto

cd ../
python -m grpc_tools.protoc -I./intrigue --python_out=. --grpc_python_out=. ./intrigue/intrigue.proto
mv *.py ./intrigue/

## Python Deps
# sudo python -m pip install grpcio
# sudo python -m pip install grpcio-tools
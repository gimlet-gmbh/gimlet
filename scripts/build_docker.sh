#!/bin/bash

echo "Building Docker"
docker rmi -f gmbh-img
docker build -t gmbh-img ./
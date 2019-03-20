#!/bin/bash

image=gmbh-img
container=gmbh-core

docker stop $container
docker rm $container
docker run -p 59500:59500 -p 49500:49500 --name $container $image
# docker run --network=host --name $container $image
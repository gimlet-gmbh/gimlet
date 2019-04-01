package config

// ProcMDkr procm dockerfile template
const ProcMDkr = `FROM golang:1.11-alpine

RUN apk add wget supervisor curl git mercurial make

## Install Deps
RUN go get github.com/golang/protobuf/proto \
    && go get github.com/golang/protobuf/protoc-gen-go  \
    && go get google.golang.org/grpc \
    && go get github.com/fatih/color \
    && go get github.com/BurntSushi/toml \
    && go get github.com/rs/xid


ENV SRCDIR=/build/gmbh
RUN mkdir -p $SRCDIR

WORKDIR $SRCDIR

ARG CACHEBUST=1

RUN git clone https://github.com/gmbh-micro/gmbh.git \ 
    && cd gmbh \
    && git checkout containerBuild \
    && git fetch \
    && mkdir -p $GOPATH"/src/github.com/gmbh-micro" \ 
    && cp -a ./internal/* $GOPATH"/src/github.com/gmbh-micro"/ 

WORKDIR $SRCDIR/gmbh

RUN go build -v -o ./bin/gmbh ./cmd/gmbh/*.go \
    && go build -v -o ./bin/gmbhCore ./cmd/gmbhCore/*.go \
    && go build -v -o ./bin/gmbhProcm ./cmd/gmbhProcm/*.go \
    && cp ./bin/gmbh* $GOPATH/bin

WORKDIR /

CMD ["gmbhProcm"]`

// CoreDkr MUST SPECIFY THE CONFIG FILE USING SPRINTF
const CoreDkr = `FROM golang:1.11-alpine
RUN apk add wget supervisor curl git mercurial make

## Install Deps
RUN go get github.com/golang/protobuf/proto \
    && go get github.com/golang/protobuf/protoc-gen-go  \
    && go get google.golang.org/grpc \
    && go get github.com/fatih/color \
    && go get github.com/BurntSushi/toml \
    && go get github.com/rs/xid


ENV SRCDIR=/build/gmbh
RUN mkdir -p $SRCDIR

WORKDIR $SRCDIR

ARG CACHEBUST=1

RUN git clone https://github.com/gmbh-micro/gmbh.git \ 
    && cd gmbh \
    && git checkout containerBuild \
    && git fetch \
    && mkdir -p $GOPATH"/src/github.com/gmbh-micro" \ 
    && cp -a ./internal/* $GOPATH"/src/github.com/gmbh-micro"/ \
    && cp -a ./pkg/* $GOPATH"/src/github.com/gmbh-micro"/

WORKDIR $SRCDIR/gmbh

RUN go build -v -o ./bin/gmbh ./cmd/gmbh/*.go \
    && go build -v -o ./bin/gmbhCore ./cmd/gmbhCore/*.go \
    && go build -v -o ./bin/gmbhProcm ./cmd/gmbhProcm/*.go \
    && cp ./bin/gmbh* $GOPATH/bin

WORKDIR /

ADD ./gmbh-deploy/core.toml ./
ADD %s ./configFile.toml

CMD ["gmbhProcm", "--remote", "--config=./core.toml", "--verbose"]`

// ServiceDkr docker template
const ServiceDkr = `FROM golang:1.11-alpine

RUN apk add wget supervisor curl git mercurial make

## Install Deps
RUN go get github.com/golang/protobuf/proto \
    && go get github.com/golang/protobuf/protoc-gen-go  \
    && go get google.golang.org/grpc \
    && go get github.com/fatih/color \
    && go get github.com/BurntSushi/toml \
    && go get github.com/rs/xid


ENV SRCDIR=/build/gmbh
ENV SERVICEDIR=/services
RUN mkdir -p $SRCDIR; mkdir -p $SERVICEDIR 

WORKDIR $SRCDIR

ARG CACHEBUST=1

RUN git clone https://github.com/gmbh-micro/gmbh.git \ 
    && cd gmbh \
    && git checkout containerBuild \
    && git fetch \
    && mkdir -p $GOPATH"/src/github.com/gmbh-micro" \ 
    && cp -a ./internal/* $GOPATH"/src/github.com/gmbh-micro"/ \
    && cp -a ./pkg/* $GOPATH"/src/github.com/gmbh-micro"/

WORKDIR $SERVICEDIR

# ADD ./services/c0 ./c0
%s

## INSTRUCTIONS FOR BUILDING SERVICES
%s

WORKDIR $SRCDIR/gmbh

RUN go build -v -o ./bin/gmbhProcm ./cmd/gmbhProcm/*.go \
    && cp ./bin/gmbh* $GOPATH/bin

WORKDIR /

ADD ./gmbh-deploy/%s.toml ./

CMD ["gmbhProcm", "--remote", "--config=./%s.toml", "--verbose"]`

const Bash = `#!/bin/bash

check_error() {
    if [[ $1 != 0 ]]; then 
        echo "error building docker container, check dockerfiles or regenerate gmbh deployment"
        exit 1  
    fi
}
`

const CheckBash = `check_error $?
`

const Compose = `
version: '3.3'
services:
  node_procm:
    image: "gmbh-img-procm"
    environment:
      - "HOSTNAME=node_procm"
    env_file:
      - gmbh.env

  node_0:
    image: "gmbh-img-core"
    environment:
      - "HOSTNAME=node_0"
    env_file:
      - gmbh.env

  dashboard:
    image: "gmbh-dashboard-image"
    ports:
      - "5001:5001"
    logging:
      driver: "none"
`

const ComposeNode = `
  node_%d:
    image: "gmbh-img-node_%d"%s
    environment:
      - "HOSTNAME=node_%d"
    env_file:
      - gmbh.env
`

const EnvFile = `ENV=C
PROCM=node_procm:59500
CORE=node_0:49500
`

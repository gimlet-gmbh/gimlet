@echo off

echo Installing Dependencies:

echo Beginning github.com/golang/protobuf/proto
go get -u github.com/golang/protobuf/proto
echo Installed github.com/golang/protobuf/proto

echo Beginning github.com/golang/protobuf/protoc-gen-go
go get -u github.com/golang/protobuf/protoc-gen-go
echo Installed github.com/golang/protobuf/protoc-gen-go

echo Beginning google.golang.org/grpc
go get -u google.golang.org/grpc
echo Installed google.golang.org/grpc

echo Beginning github.com/fatih/color
go get -u github.com/fatih/color
echo Installed github.com/fatih/color

echo Beginning github.com/BurntSushi/toml
go get -u github.com/BurntSushi/toml
echo Installed github.com/BurntSushi/toml

echo Beginning github.com/rs/xid
go get -u github.com/rs/xid
echo Installed github.com/rs/xid
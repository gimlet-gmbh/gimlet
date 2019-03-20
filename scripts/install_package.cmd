@echo off

cd ..

echo Building GMBH Client
go build -o ./bin/gmbh ./cmd/gmbh/
copy /y .\bin\gmbh "%GOPATH%"\bin
echo Building GMBH Core
go build -o ./bin/gmbhCore ./cmd/gmbhCore/
copy /y .\bin\gmbhCore "%GOPATH%"\bin
echo Building GMBH Procm
go build -o ./bin/gmbhProcm ./cmd/gmbhProcm/
copy /y .\bin\gmbhProcm "%GOPATH%"\bin

cd scripts
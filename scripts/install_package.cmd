@echo off

cd ..

echo Building GMBH Client
go build -o ./bin/gmbh.exe ./cmd/gmbh/
copy /y .\bin\gmbh.exe "%GOPATH%"\bin
echo Building GMBH Core
go build -o ./bin/gmbhCore.exe ./cmd/gmbhCore/
copy /y .\bin\gmbhCore.exe "%GOPATH%"\bin
echo Building GMBH Procm
go build -o ./bin/gmbhProcm.exe ./cmd/gmbhProcm/
copy /y .\bin\gmbhProcm.exe "%GOPATH%"\bin

cd scripts
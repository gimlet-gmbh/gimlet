@ECHO off

IF "%GOPATH%"=="" (
    ECHO "GOPATH must be set before installing gmbh packages."
    ECHO "To set to the default GOPATH add \`export GOPATH=\$HOME/go\` to your bash profile"
    EXIT /B
)

ECHO Uninstalling old versions of gmbh-micro
rd /s "%GOPATH%/src/github.com/gmbh-micro"
rd /s "%GOPATH%/src/github.com/gimlet-gmbh"

@ECHO off

IF "%GOPATH%"=="" (
    ECHO "GOPATH must be set before installing gmbh packages."
    ECHO "To set to the default GOPATH add \`export GOPATH=\$HOME/go\` to your bash profile"
    EXIT /B
)

call link_packagesWin.cmd
call install_deps.cmd
call install_package.cmd

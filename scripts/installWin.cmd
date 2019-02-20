@ECHO off

IF "%GOPATH%"=="" (
    ECHO "GOPATH must be set before installing gmbh packages."
    ECHO "To set to the default GOPATH add \`export GOPATH=\$HOME/go\` to your bash profile"
    EXIT /B
)

link_packagesWin.cmd
cd ..
install_deps.cmd
install_package.cmd

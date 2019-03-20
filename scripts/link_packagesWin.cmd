@ECHO off
SETLOCAL

IF ("%GOPATH%" == "") (
    ECHO "GOPATH must be set before installing gmbh packages."
    EXIT /B
)

:: Link packages needed to build gmbh server and Go package
SET GMBH_PATH=%GOPATH%\src\github.com\gmbh-micro
mkdir "%GMBH_PATH%"

ECHO linking gmbh at %GMBH_PATH%

cd ../internal/

:: Link the internal packages for building the core
:: ECHO Linking: %%P & SET pdir=%cd%\%%P & mklink /D "%pdir%" "%GMBH_PATH%"
FOR /D %%P in ("*") DO IF NOT EXIST "%GMBH_PATH%\%%P" ( ECHO Linking: %%P & mklink /D "%GMBH_PATH%\%%P" "%cd%\%%P" ) ELSE ( ECHO "%%P is already linked" )

cd ../pkg/

:: Link the go client package
IF EXIST ("%GMBH_PATH%\gmbh") (
    ECHO "gmbh is already linked"
) ELSE (
    mklink /D "%GMBH_PATH%\gmbh" "%cd%\gmbh"
    ECHO linking gmbh to go path
)

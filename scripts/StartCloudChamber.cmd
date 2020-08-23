@rem
@rem
@rem    S T A R T C L O U D C H A M B E R . C M D
@rem
@rem
@rem

@if /i "%DbgScript%" == "" @echo off

SETLOCAL ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION

set SCRIPTDIR=%~dp0
set CLOUDCHAMBERDIR=%SCRIPTDIR:~0,-7%
set CLOUDCHAMBERFILE=%CLOUDCHAMBERDIR%\Files
set CLOUDCHAMBERDATA=%CLOUDCHAMBERDIR%\Data



rem Check for requests for help without actually doing anything
rem 
if /i "%1" == "/?"     (goto :ScriptHelp)
if /i "%1" == "-?"     (goto :ScriptHelp)
if /i "%1" == "/h"     (goto :ScriptHelp)
if /i "%1" == "-h"     (goto :ScriptHelp)
if /i "%1" == "--help" (goto :ScriptHelp)



rem To allow for the main config file cloudchamber.yaml to be location independant, all
rem included paths are relative rather than absolute. This requires that we set the 
rem current (working) directory to math the expectations of the config file.
rem
rem If at some point we can feed in the paths as arguments, this restriction can be relaxed
rem 

pushd %CLOUDCHAMBERFILE%
call :StartBinary controllerd.exe  %CLOUDCHAMBERFILE% %CLOUDCHAMBERFILE% %GOPATH%\bin
call :StartBinary inventoryd.exe   %CLOUDCHAMBERFILE% %CLOUDCHAMBERFILE% %GOPATH%\bin
call :StartBinary sim_supportd.exe %CLOUDCHAMBERFILE% %CLOUDCHAMBERFILE% %GOPATH%\bin
call :StartBinary web_server.exe   %CLOUDCHAMBERFILE% %CLOUDCHAMBERFILE% %GOPATH%\bin
popd

goto :ScriptExit




rem Starts the binary in %1 using config in %2
rem
rem Searches paths %3, %4, PATH in that order.
rem 

:StartBinary

rem Find a binary to use
rem
if exist %3\%1 (

  set TARGETBIN=%3\%1

) else if exist %4\%1 (

  set TARGETBIN=%4\%1

) else (

   for %%I in (%BINARY%) do set TARGETBIN=%%~$PATH:I

)


if not exist "%TARGETBIN%" (
  echo.
  echo Unable to find a copy of %1
  echo.
  goto :StartBinaryExit
)


rem Now actually start the binary
rem
echo.
echo Starting %TARGETBIN%

start %TARGETBIN% -config=%2

goto :StartBinaryExit


:StartBinaryExit

goto :EOF





:ScriptHelp

echo.
echo StartCloudChamber
echo.
echo Starts an instance of the CloudChamber services using the configuration file.
echo.
echo The service binaries are expected to be located either in a
echo standard deployment directory or as a fallback, from the
echo %GOPATH%\bin directory. The deployment directory is located
echo based on the location of the StartCloudChamber.cmd script
echo itself.
echo.

goto :ScriptExit



:ScriptExit

ENDLOCAL
goto :EOF

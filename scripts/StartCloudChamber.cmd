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
set CLOUDCHAMBERLOGS=%CLOUDCHAMBERDIR%\Logs


rem Check for requests for help without actually doing anything
rem 
if /i "%1" == "/?"     (goto :ScriptHelp)
if /i "%1" == "-?"     (goto :ScriptHelp)
if /i "%1" == "/h"     (goto :ScriptHelp)
if /i "%1" == "-h"     (goto :ScriptHelp)
if /i "%1" == "--help" (goto :ScriptHelp)



set UpdateDate=%date%
set UpdateTime=%time%

set UpdateYear=%UpdateDate:~10,4%
set UpdateDay=%UpdateDate:~7,2%
set UpdateMonth=%UpdateDate:~4,2%

set UpdateHour=%UpdateTime:~0,2%
set UpdateMinute=%UpdateTime:~3,2%
set UpdateSecond=%UpdateTime:~6,2%


rem Allow for some variants dumping the time var with a leading space rather than a leading zero.
rem

if " " == "%UpdateHour:~0,1%" set UpdateHour=0%UpdateHour:~1,1%

set UpdateDateTime=%UpdateYear%%UpdateMonth%%UpdateDay%-%UpdateHour%%UpdateMinute%%UpdateSecond%


if not exist "%CLOUDCHAMBERLOGS%" (mkdir "%CLOUDCHAMBERLOGS%")


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

set BINARY=%1

if exist %3\%BINARY% (

  set TARGETBIN=%3\%BINARY%

) else if exist %4\%BINARY% (

  set TARGETBIN=%4\%BINARY%

) else (

   for %%I in (%BINARY%) do set TARGETBIN=%%~$PATH:I

)


if not exist "%TARGETBIN%" (
  echo.
  echo Unable to find a copy of %BINARY%
  echo.
  goto :StartBinaryExit
)


rem Now actually start the binary
rem
echo.
echo Starting %TARGETBIN%

start cmd /c "%TARGETBIN% -config=%2 2>&1 >%CLOUDCHAMBERLOGS%\%BINARY:~0,-4%.log"
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

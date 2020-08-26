@rem
@rem
@rem    M O N I T O R E T C D . C M D
@rem
@rem
@rem

@if /i "%DbgScript%" == "" @echo off

SETLOCAL ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION

set SCRIPTDIR=%~dp0
set CLOUDCHAMBERDIR=%SCRIPTDIR:~0,-7%
set CLOUDCHAMBERFILE=%CLOUDCHAMBERDIR%\Files
set CLOUDCHAMBERDATA=%CLOUDCHAMBERDIR%\Data


set LOCALHOST=127.0.0.1

set BINARY=etcdctl.exe
set TARGETBINPATH=%ETCDCTLBINPATH%

set DEFAULT_ETCDINSTANCE=%COMPUTERNAME%
set DEFAULT_ETCDNODEADDR=%LOCALHOST%
set DEFAULT_ETCDPORTCLNT=2379


rem Setup the necessary variables if not overriden
rem 

if /i "%ETCDNODEADDR%" == "" (set ETCDNODEADDR=%DEFAULT_ETCDNODEADDR%)
if /i "%ETCDPORTCLNT%" == "" (set ETCDPORTCLNT=%DEFAULT_ETCDPORTCLNT%)


rem Check for requests for help without actually doing anything
rem 
if /i "%1" == "/?"     (goto :ScriptHelp)
if /i "%1" == "-?"     (goto :ScriptHelp)
if /i "%1" == "/h"     (goto :ScriptHelp)
if /i "%1" == "-h"     (goto :ScriptHelp)
if /i "%1" == "--help" (goto :ScriptHelp)


rem Find a binary to use
rem
if exist %CLOUDCHAMBERFILE%\etcdctl.exe (

  set TARGETBIN=%CLOUDCHAMBERFILE%\etcdctl.exe

) else if exist %TARGETBINPATH%\etcdctl.exe (

  set TARGETBIN=%TARGETBINPATH%\etcdctl.exe

) else if exist %GOPATH%\bin\etcdctl.exe (

  set TARGETBIN=%GOPATH%\bin\etcdctl.exe

) else (

   for %%I in (etcdctl.exe) do set TARGETBIN=%%~$PATH:I

)


if not exist "%TARGETBIN%" (
  echo.
  echo Unable to find a copy of etcdctl.exe
  echo.
  goto :ScriptExit
)


rem Now actually start the utility
rem
echo.
echo Using %TARGETBIN% monitoring %ETCDNODEADDR%:%ETCDPORTCLNT%

start %TARGETBIN% --endpoints=%ETCDNODEADDR%:%ETCDPORTCLNT% watch --prefix /CloudChamber/v0.1

goto :ScriptExit



:ScriptHelp

echo.
echo MonitorEtcd
echo.
echo Starts a single etcdctl session to monitor an etcd instance.
echo.
echo There are a number of (required on Windows) parameters which have default values as listed below but
echo which can be overridden by setting environment variables using the appropriate names along with the
echo desired values.
echo.
echo ETCDNODEADDR - IP address of the ETCD instance                       (defaults to %DEFAULT_ETCDNODEADDR%)
echo ETCDPORTCLNT - IP port to be used for communication with the client  (defaults to %DEFAULT_ETCDPORTCLNT%) 
echo.
echo.

goto :ScriptExit



:ScriptExit

ENDLOCAL
goto :EOF

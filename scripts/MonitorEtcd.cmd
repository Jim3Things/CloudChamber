@rem
@rem
@rem    M O N I T O R E T C D . C M D
@rem
@rem
@rem

@if /i "%DbgScript%" == "" @echo off

SETLOCAL ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION

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
if exist %~dp0%BINARY% (

  set TARGETBIN=%~dp0%BINARY%

) else if exist %TARGETBINPATH%\%BINARY% (

  set TARGETBIN=%TARGETBINPATH%\%BINARY%

) else if exist %GOPATH%\bin\%BINARY% (

  set TARGETBIN=%GOPATH%\bin\%BINARY%

) else (

   for %%I in (%BINARY%) do set TARGETBIN=%%~$PATH:I

)


if not exist "%TARGETBIN%" (
  echo.
  echo Unable to find a copy of %BINARY%
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

echo StartEtcd
echo.
echo Starts a single etcdctl session to monitor an etcd instance.
echo.
echo There are a number of (required on Windows) parameters which have default values as listed below but
echo which can be overridden by setting environment variables using the appropriate names along with the
echo desired values.
echo.
echo ETCDNODEADDR (defaults to %DEFAULT_ETCDNODEADDR%) - IP address of the ETCD instance
echo ETCDPORTCLNT (defaults to %DEFAULT_ETCDPORTCLNT%) - IP port to be used for communication with the client
echo.
echo.

goto :ScriptExit



:ScriptExit

ENDLOCAL
goto :EOF

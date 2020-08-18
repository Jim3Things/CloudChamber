@rem
@rem
@rem    S T A R T E T C D . C M D
@rem
@rem
@rem

@if /i "%DbgScript%" == "" @echo off

SETLOCAL ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION

set LOCALHOST=127.0.0.1

set BINARY=etcd.exe
set TARGETBINPATH=%ETCDBINPATH%

set DEFAULT_ETCDROOT=%SystemDrive%\etcd
set DEFAULT_ETCDDATA=%DEFAULT_ETCDROOT%\data
set DEFAULT_ETCDINSTANCE=%COMPUTERNAME%
set DEFAULT_ETCDNODEADDR=%LOCALHOST%
set DEFAULT_ETCDPORTCLNT=2379
set DEFAULT_ETCDPORTPEER=2380


rem Setup the necessary variables if not overriden
rem 

if /i "%ETCDINSTANCE%" == "" (set ETCDINSTANCE=%DEFAULT_ETCDINSTANCE%)
if /i "%ETCDNODEADDR%" == "" (set ETCDNODEADDR=%DEFAULT_ETCDNODEADDR%)
if /i "%ETCDPORTCLNT%" == "" (set ETCDPORTCLNT=%DEFAULT_ETCDPORTCLNT%)
if /i "%ETCDPORTPEER%" == "" (set ETCDPORTPEER=%DEFAULT_ETCDPORTPEER%)


rem Check for requests for help without actually doing anything
rem 
if /i "%1" == "/?"     (goto :ScriptHelp)
if /i "%1" == "-?"     (goto :ScriptHelp)
if /i "%1" == "/h"     (goto :ScriptHelp)
if /i "%1" == "-h"     (goto :ScriptHelp)
if /i "%1" == "--help" (goto :ScriptHelp)


rem Decide on a path to the data
rem
if /i "%1" NEQ "" (

  set ETCDDIR=%1

) else if "%ETCDDATA%" == "" (

  set ETCDDIR=%DEFAULT_ETCDDATA%\%ETCDINSTANCE%.etcd

) else (

  set ETCDDIR=%ETCDDATA%\%ETCDINSTANCE%.etcd

)


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


rem Ensure the root data directory is present
rem
if not exist "%ETCDDIR%" (mkdir "%ETCDDIR%")


rem Now actually start the ETCD service
rem
echo.
echo Using %TARGETBIN% writing to "%ETCDDIR%"

start %TARGETBIN% --name "%ETCDINSTANCE%" --data-dir "%ETCDDIR%" --listen-peer-urls "http://%LOCALHOST%:%ETCDPORTPEER%" --listen-client-urls "http://%LOCALHOST%:%ETCDPORTCLNT%" --advertise-client-urls "http://%LOCALHOST%:%ETCDPORTCLNT%"

goto :ScriptExit



:ScriptHelp

echo StartEtcd
echo.
echo Starts a single etcd instance.
echo.
echo There are a number of (required on Windows) parameters which have default values as listed below but
echo which can be overridden by setting environment variables using the appropriate names along with the
echo desired values.
echo.
echo ETCDINSTANCE (defaults to %DEFAULT_ETCDINSTANCE%) - name of the ETCD instance
echo ETCDNODEADDR (defaults to %DEFAULT_ETCDNODEADDR%) - IP address of the ETCD instance
echo ETCDPORTCLNT (defaults to %DEFAULT_ETCDPORTCLNT%) - IP port to be used for communication with the client
echo ETCDDATA     (defaults to %DEFAULT_ETCDDATA%)     - directory where the ETCD data files are to be placed
echo.
echo.

goto :ScriptExit



:ScriptExit

ENDLOCAL
goto :EOF

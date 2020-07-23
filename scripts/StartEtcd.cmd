@rem
@rem
@rem    S T A R T E T C D . C M D
@rem
@rem
@rem

@if /i "%DbgScript%" == "" @echo off

SETLOCAL ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION

set LOCALHOST=127.0.0.1

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
if /i "%1" == "/?"     (goto :StartEtcdHelp)
if /i "%1" == "-?"     (goto :StartEtcdHelp)
if /i "%1" == "/h"     (goto :StartEtcdHelp)
if /i "%1" == "-h"     (goto :StartEtcdHelp)
if /i "%1" == "--help" (goto :StartEtcdHelp)


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
if exist %ETCDBINPATH%\etcd.exe (

  set ETCDBIN=%ETCDBINPATH%\etcd.exe

) else if exist %GOPATH%\bin\etcd.exe (

  set ETCDBIN=%GOPATH%\bin\etcd.exe

) else (
   for %%I in (etcd.exe) do set ETCDBIN=%%~$PATH:I
)


if not exist "%ETCDBIN%" (
echo.
echo Unable to find a copy of etcd.exe
echo.
goto :StartEtcdExit
)


rem Ensure the root data directory is present
rem
if not exist "%ETCDDIR%" (mkdir "%ETCDDIR%")


rem Now actually start the ETCD service
rem
echo.
echo Using %ETCDBIN% writing to "%ETCDDIR%"

start %ETCDBIN% --name "%ETCDINSTANCE%" --data-dir "%ETCDDIR%" --listen-peer-urls "http://%LOCALHOST%:%ETCDPORTPEER%" --listen-client-urls "http://%LOCALHOST%:%ETCDPORTCLNT%" --advertise-client-urls "http://%LOCALHOST%:%ETCDPORTCLNT%"

goto :StartEtcdExit



:StartEtcdHelp

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

goto :StartEtcdExit



:StartEtcdExit

ENDLOCAL
goto :EOF

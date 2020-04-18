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

if /i "%ETCDDATA%" == "" (set ETCDDATA=%DEFAULT_ETCDDATA%)

if /i "%ETCDINSTANCE%" == "" (set ETCDINSTANCE=%DEFAULT_ETCDINSTANCE%)
if /i "%ETCDPORTCLNT%" == "" (set ETCDPORTCLNT=%DEFAULT_ETCDPORTCLNT%)
if /i "%ETCDPORTPEER%" == "" (set ETCDPORTPEER=%DEFAULT_ETCDPORTPEER%)


rem Check for requests for help without actually doing anything
rem 
if /i "%1" == "/?" (goto :StartEtcdHelp)
if /i "%1" == "/h" (goto :StartEtcdHelp)
if /i "%1" == "-?" (goto :StartEtcdHelp)
if /i "%1" == "-h" (goto :StartEtcdHelp)



rem Ensure the root data directory is present
rem
if not exist "%ETCDDATA%" (mkdir "%ETCDDATA%")


rem Now actually start the ETCD service
rem
start %GOPATH%\bin\etcd --name "%ETCDINSTANCE%" --data-dir "%ETCDDATA%\%ETCDINSTANCE%.etcd" --listen-peer-urls "http://%LOCALHOST%:%ETCDPORTPEER%" --listen-client-urls "http://%LOCALHOST%:%ETCDPORTCLNT%" --advertise-client-urls "http://%LOCALHOST%:%ETCDPORTCLNT%"

goto :StartEtcdExit



:StartEtcdHelp

echo StartEtcd
echo.
echo Starts a single etcd instance.
echo.
echo There are a number of (required on Windows) parameters which have default values as listed below but
echo which can be overriden by setting environment variables using the appropriate names along with the
echo desired values.
echo.
echo ETCDINSTANCE (defaults to %DEFAULT_ETCDINSTANCE%) - name of the ETCD instance
echo ETCDPORTCLNT (defaults to %DEFAULT_ETCDPORTCLNT%) - IP port to be used for communication with the client
echo ETCDDATA     (defaults to %DEFAULT_ETCDDATA%)     - directory where the ETCD data files are to be placed- 
echo.
echo.

goto :StartEtcdExit



:StartEtcdExit

ENDLOCAL
goto :EOF

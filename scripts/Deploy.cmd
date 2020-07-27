@rem
@rem
@rem    D E P L O Y . C M D
@rem
@rem
@rem

@if /i "%DbgScript%" == "" @echo off

SETLOCAL ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION

set DEFAULT_PARAM_VALUE_DEPLOYMENT_DIR=%SystemDrive%\CloudChamber
set DEFAULT_PARAM_VALUE_ETCD=Include



rem Parameters
rem
set DEPLOY_PARAM_NAME_TARGETDIR=-TargetDir
set DEPLOY_PARAM_NAME_ETCD=-Etcd
set DEPLOY_PARAM_NAME_NOETCD=-NoEtcd


rem -NoEtcd
rem -TargetDir
rem -Package

set DeployTargetDir=
set DeployEtcd=
set DeployPackage=


:DeployParseLoopStart

rem Check for requests for help without actually doing anything
rem 
if /i "%1" == ""       (goto :DeployHelp)
if /i "%1" == "/?"     (goto :DeployHelp)
if /i "%1" == "-?"     (goto :DeployHelp)
if /i "%1" == "/h"     (goto :DeployHelp)
if /i "%1" == "-h"     (goto :DeployHelp)
if /i "%1" == "--help" (goto :DeployHelp)


if /i "%1" == "DEPLOY_PARAM_NAME_TARGETDIR" (

  shift

  if /i "%1" == "" (goto :DeployHelp)

  set DeployTargetDir=%2
  shift

  goto :DeployParseLoopStart

) else if /i "%1" == "DEPLOY_PARAM_NAME_ETCD" (

  set DeployEtcd=Include
  shift

  goto :DeployParseLoopStart

) else if /i "%1" == "DEPLOY_PARAM_NAME_NOETCD" (

  set DeployEtcd=Exclude
  shift

  goto :DeployParseLoopStart

) else if /i "%1" == "-Package" (

  shift
  
  goto :DeployParseLoopStart

) else if /i "%1" != "" (

  goto :DeployParseLoopStart

)




rem Decide on a path to the root to the deployment
rem
if /i "%DeployTargetDir%" NEQ "" (

  set CLOUDCHAMBERDIR=%DeployTargetDir%

) else if /i "%CLOUDCHAMBER%" NEQ "" (

  set CLOUDCHAMBERDIR=%CLOUDCHAMBER%

) else (

  set CLOUDCHAMBERDIR=%DEFAULT_PARAM_VALUE_DEPLOYMENT_DIR%

)


rem Decide on a whether or not the etcd.exe and etcdutl.exe binaries should be included
rem

if /i "%DeployEtcd%" == "" (

  set DeployEtcd=%DEFAULT_PARAM_VALUE_ETCD%

)



xcopy /e /r /h /k %GOPATH%\src\github.com\Jim3Things\cloud_chamber_react_ts\build\*             %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments\cloudchamber.yaml %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments\controllerd.exe   %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments\inventoryd.exe    %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments\sim_supportd.exe  %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments\web_server.exe    %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\scripts\StartAll.cmd          %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\scripts\StartCloudChamber.cmd %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\scripts\StartEtcd.cmd         %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\scripts\MonitorEtcd.cmd       %CloudChamber%\Files\


if /i "%DeployEtcd%" EQU "Include" (

  xcopy %GOPATH%\bin\etcd.exe    %CloudChamber%\Files\
  xcopy %GOPATH%\bin\etcdutl.exe %CloudChamber%\Files\

  echo %CloudChamber%\Data      >%CloudChamber%\Files\EtcDataDir.config

)

goto :DeployExit


:DeployHelp

echo Deploy
echo.
echo Deploys a copy of Cloudchamber to the installation directory
echo.
echo %DEPLOY_PARAM_NAME_TARGETDIR%   (defaults to %DEFAULT_PARAM_VALUE_DEPLOYMENT_DIR%)
echo %DEPLOY_PARAM_NAME_ETCD%        (defaults to %DEFAULT_PARAM_VALUE_ETCD%

rem -NoEtcd
rem -TargetDir

:DeployExit

ENDLOCAL
goto :EOF

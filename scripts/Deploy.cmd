@rem
@rem
@rem    D E P L O Y . C M D
@rem
@rem
@rem

@if /i "%DbgScript%" == "" @echo off

SETLOCAL ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION


set CLOUDCHAMBER_KIT=%GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments
set CLOUDCHAMBER_UI=%GOPATH%\src\github.com\Jim3Things\cloud_chamber_react_ts\build


set DEFAULT_PARAM_VALUE_DEPLOYMENT_DIR=%SystemDrive%\CloudChamber
set DEFAULT_PARAM_VALUE_ETCD=Exclude



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


if /i "%1" == "%DEPLOY_PARAM_NAME_TARGETDIR%" (

  shift

  if /i "%1" == "" (goto :DeployHelp)

  set DeployTargetDir=%2
  shift

  goto :DeployParseLoopStart

) else if /i "%1" == "%DEPLOY_PARAM_NAME_ETCD%" (

  set DeployEtcd=Include
  shift

  goto :DeployParseLoopStart

) else if /i "%1" == "%DEPLOY_PARAM_NAME_NOETCD%" (

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



xcopy /e /r /h /k %CLOUDCHAMBER_UI%\*                      %CloudChamberDir%\Files\
xcopy             %CLOUDCHAMBER_KIT%\cloudchamber.yaml     %CloudChamberDir%\Files\
xcopy             %CLOUDCHAMBER_KIT%\controllerd.exe       %CloudChamberDir%\Files\
xcopy             %CLOUDCHAMBER_KIT%\inventoryd.exe        %CloudChamberDir%\Files\
xcopy             %CLOUDCHAMBER_KIT%\sim_supportd.exe      %CloudChamberDir%\Files\
xcopy             %CLOUDCHAMBER_KIT%\web_server.exe        %CloudChamberDir%\Files\
xcopy             %CLOUDCHAMBER_KIT%\StartAll.cmd          %CloudChamberDir%\Files\
xcopy             %CLOUDCHAMBER_KIT%\StartCloudChamber.cmd %CloudChamberDir%\Files\
xcopy             %CLOUDCHAMBER_KIT%\StartEtcd.cmd         %CloudChamberDir%\Files\
xcopy             %CLOUDCHAMBER_KIT%\MonitorEtcd.cmd       %CloudChamberDir%\Files\


if /i "%DeployEtcd%" EQU "Include" (


  call :CopyEtcdBin etcd.exe    %CloudChamberDir%\Files
  call :CopyEtcdBin etcdctl.exe %CloudChamberDir%\Files

  echo %CloudChamberDir%\Data  >%CloudChamberDir%\Files\EtcDataDir.config
)

goto :DeployExit




:DeployHelp

echo Deploy
echo.
echo Deploys a copy of Cloudchamber to the installation directory
echo.
echo %DEPLOY_PARAM_NAME_TARGETDIR%        (defaults to %DEFAULT_PARAM_VALUE_DEPLOYMENT_DIR%)
echo %DEPLOY_PARAM_NAME_ETCD% | %DEPLOY_PARAM_NAME_NOETCD%    (defaults to %DEFAULT_PARAM_VALUE_ETCD%

goto :DeployExit




rem Find a binary to use from one of three locations defined by the environment variables
rem
rem  - ETCDBINPATH
rem  - GOPATH
rem  - PATH
rem

:CopyEtcdBin

if exist %ETCDBINPATH%\%1 (

  set TARGETBIN=%ETCDBINPATH%\%1

) else if exist %GOPATH%\bin\%1 (

  set TARGETBIN=%GOPATH%\bin\%1

) else (

  for %%I in (%1) do set TARGETBIN=%%~$PATH:I

)


if not exist "%TARGETBIN%" (
  echo.
  echo Unable to find a version of %1 to copy
  echo.
  goto :CopyEtcdBinExit
)


xcopy %TARGETBIN% %2\



:CopyEtcdBinExit

goto :EOF




:DeployExit

ENDLOCAL
goto :EOF


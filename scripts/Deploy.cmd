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

set DEPLOY_PARAM_VALUE_ETCD_INCLUDE=Include
set DEPLOY_PARAM_VALUE_ETCD_EXCLUDE=Exclude



set DEFAULT_PARAM_VALUE_DEPLOYMENT_DIR=%SystemDrive%\CloudChamber
set DEFAULT_PARAM_VALUE_ETCD=%DEPLOY_PARAM_VALUE_ETCD_EXCLUDE%


rem Parameters
rem
set DEPLOY_PARAM_NAME_TARGETDIR=-TargetDir
set DEPLOY_PARAM_NAME_ETCD=-Etcd
set DEPLOY_PARAM_NAME_NOETCD=-NoEtcd


set DeployTargetDir=
set DeployEtcd=
set DeployPackage=


rem Check for requests for help without actually doing anything
rem 
if /i "%1" == ""       (goto :ScriptHelp)


:DeployParseLoopStart

if /i "%1" == "/?"     (goto :ScriptHelp)
if /i "%1" == "-?"     (goto :ScriptHelp)
if /i "%1" == "/h"     (goto :ScriptHelp)
if /i "%1" == "-h"     (goto :ScriptHelp)
if /i "%1" == "--help" (goto :ScriptHelp)


if /i "%1" == "%DEPLOY_PARAM_NAME_TARGETDIR%" (

  if /i "%2" == "" (goto :ScriptHelp)

  set DeployTargetDir=%2
  shift
  shift

  goto :DeployParseLoopStart

) else if /i "%1" == "%DEPLOY_PARAM_NAME_ETCD%" (

  set DeployEtcd=%DEPLOY_PARAM_VALUE_ETCD_INCLUDE%
  shift

  goto :DeployParseLoopStart

) else if /i "%1" == "%DEPLOY_PARAM_NAME_NOETCD%" (

  set DeployEtcd=Exclude
  shift

  goto :DeployParseLoopStart

) else if /i "%1" == "-Package" (

  shift
  
  goto :DeployParseLoopStart

) else if /i "%1" NEQ "" (

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



xcopy /e /r /h /k %CLOUDCHAMBER_UI%\*  %CloudChamberDir%\Files\
xcopy             %CLOUDCHAMBER_KIT%\* %CloudChamberDir%\Files\


if /i "%DeployEtcd%" EQU "%DEPLOY_PARAM_VALUE_ETCD_INCLUDE%" (

  call :CopyEtcdBin etcd.exe    %CloudChamberDir%\Files
  call :CopyEtcdBin etcdctl.exe %CloudChamberDir%\Files
)

goto :ScriptExit




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




:ScriptHelp

echo.
echo Deploy [%DEPLOY_PARAM_NAME_TARGETDIR% ^<path^>] [%DEPLOY_PARAM_NAME_ETCD% ^| %DEPLOY_PARAM_NAME_NOETCD%]
echo.
echo Deploys a copy of Cloudchamber to the specified installation
echo directory using files from
echo.
echo   %CLOUDCHAMBER_KIT%
echo   %CLOUDCHAMBER_UI%
echo.
echo.
echo   %DEPLOY_PARAM_NAME_TARGETDIR% ^<path^>   (defaults to %DEFAULT_PARAM_VALUE_DEPLOYMENT_DIR%)
echo.
echo    Identifies the target path for the deployment
echo.
echo   %DEPLOY_PARAM_NAME_ETCD%  ^| %DEPLOY_PARAM_NAME_NOETCD%    (defaults to %DEPLOY_PARAM_NAME_NOETCD%)
echo.
echo    Indicates whether or not a copy of the ectd binaries
echo    should be included with the deployment.
echo.

goto :ScriptExit




:ScriptExit

ENDLOCAL
goto :EOF


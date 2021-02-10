@rem
@rem
@rem    S T A R T A L L . C M D
@rem
@rem
@rem

@if /i "%DbgScript%" == "" @echo off

SETLOCAL ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION

set SCRIPTDIR=%~dp0

rem Check for requests for help without actually doing anything
rem 
if /i "%1" == "/?"     (goto :ScriptHelp)
if /i "%1" == "-?"     (goto :ScriptHelp)
if /i "%1" == "/h"     (goto :ScriptHelp)
if /i "%1" == "-h"     (goto :ScriptHelp)
if /i "%1" == "--help" (goto :ScriptHelp)

set DEFAULT_DEPLOYMENT=%SystemDrive%\CloudChamber

rem Decide on a path to the root to the deployment
rem
if /i "%1" == "" (

  set CLOUDCHAMBERDIR=%SCRIPTDIR:~0,-7%

) else if /i "%CLOUDCHAMBER%" NEQ "" (

  set CLOUDCHAMBERDIR=%1

) else (

  set CLOUDCHAMBERDIR=%DEFAULT_DEPLOYMENT%

)

set CLOUDCHAMBERFILE=%CLOUDCHAMBERDIR%\Files
set CLOUDCHAMBERDATA=%CLOUDCHAMBERDIR%\Data



call %CLOUDCHAMBERFILE%\StartEtcd.cmd %CLOUDCHAMBERDATA%
call %CLOUDCHAMBERFILE%\MonitorEtcd.cmd
call %CLOUDCHAMBERFILE%\StartCloudChamber.cmd

goto :ScriptExit


:ScriptHelp

echo.
echo StartAll [^<DeploymentPath^>]
echo.
echo Starts an complete instance of the CloudChamber services using
echo the configuration file. This includes an etcd instance.
echo.
echo The service binaries are expected to be located in a standard
echo deployment directory identified either by the supplied path or
echo determined from the location of the StartAll.cmd script itself.
echo.

goto :ScriptExit




:ScriptExit

ENDLOCAL
goto :EOF

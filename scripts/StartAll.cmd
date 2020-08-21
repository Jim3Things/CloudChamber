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
if /i "%1" == "/?"     (goto :StartAllHelp)
if /i "%1" == "-?"     (goto :StartAllHelp)
if /i "%1" == "/h"     (goto :StartAllHelp)
if /i "%1" == "-h"     (goto :StartAllHelp)
if /i "%1" == "--help" (goto :StartAllHelp)

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

:StartAllExit

ENDLOCAL
goto :EOF

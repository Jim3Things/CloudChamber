@rem
@rem
@rem    S T A R T A L L . C M D
@rem
@rem
@rem

@if /i "%DbgScript%" == "" @echo off

SETLOCAL ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION

rem Check for requests for help without actually doing anything
rem 
if /i "%1" == "/?"     (goto :StartAllHelp)
if /i "%1" == "-?"     (goto :StartAllHelp)
if /i "%1" == "/h"     (goto :StartAllHelp)
if /i "%1" == "-h"     (goto :StartAllHelp)
if /i "%1" == "--help" (goto :StartAllHelp)

set DEFAULT_DEPLOYMENT_ROOT=%SystemDrive%\CloudChamber

set DEFAULT_DEPLOYMENT=%DEFAULT_DEPLOYMENT_ROOT%\Files

rem Decide on a path to the root to the deployment
rem
if /i "%1" NEQ "" (

  set CLOUDCHAMBERDIR=%1

) else if /i "%CLOUDCHAMBER%" NEQ "" (

  set CLOUDCHAMBERDIR=%1

) else (

  set CLOUDCHAMBERDIR=%DEFAULT_DEPLOYMENT%)

)




call %CLOUDCHAMBERDIR%\StartEtcd.cmd %CLOUDCHAMBERDIR%\..\Data
call %CLOUDCHAMBERDIR%\MonitorEtcd.cmd
call %CLOUDCHAMBERDIR%\StartCloudChamber.cmd

:StartAllExit

ENDLOCAL
goto :EOF

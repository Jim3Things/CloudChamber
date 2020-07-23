@rem
@rem
@rem    D E P L O Y . C M D
@rem
@rem
@rem

@if /i "%DbgScript%" == "" @echo off

SETLOCAL ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION

rem Check for requests for help without actually doing anything
rem 
if /i "%1" == "/?"     (goto :DeployHelp)
if /i "%1" == "-?"     (goto :DeployHelp)
if /i "%1" == "/h"     (goto :DeployHelp)
if /i "%1" == "-h"     (goto :DeployHelp)
if /i "%1" == "--help" (goto :DeployHelp)


set DEFAULT_DEPLOYMENT=%SystemDrive%\CloudChamber

rem Decide on a path to the root to the deployment
rem
if /i "%1" NEQ "" (

  set CLOUDCHAMBERDIR=%1

) else if /i "%CLOUDCHAMBER%" NEQ "" (

  set CLOUDCHAMBERDIR=%1

) else (

  set CLOUDCHAMBERDIR=%DEFAULT_DEPLOYMENT%)

)


xcopy /e /r /h /k %GOPATH%\src\github.com\Jim3Things\cloud_chamber_react_ts\build\*             %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments\*                 %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\scripts\StartAll.cmd          %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\scripts\StartCloudChamber.cmd %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\scripts\StartEtcd.cmd         %CloudChamber%\Files\
xcopy             %GOPATH%\src\github.com\Jim3Things\CloudChamber\scripts\MonitorEtcd.cmd       %CloudChamber%\Files\

goto :DeployExit


:DeployHelp



:DeployExit

ENDLOCAL
goto :EOF

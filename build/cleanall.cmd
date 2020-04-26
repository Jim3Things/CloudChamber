@rem
@rem
@rem    C L E A N A L L . C M D
@rem
@rem
@rem

@if /i "%DbgScript%" == "" @echo off

if exist "%GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments" (

  pushd "%GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments"

  del *.*

  popd

) else (
  echo.
  echo Deployments directory not found. Check definition of GOPATH: %GOPATH%
  echo.
)

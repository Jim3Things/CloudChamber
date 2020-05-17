@rem
@rem
@rem T E S T . C M D
@rem
@rem Allows one or more individually selected tests to be run, one at a time
@rem

@if /i "%DbgScript%" == "" @echo off

setlocal

if "%1" == "" goto :TestHelp


:TestParseLoop

if "%1" == "" goto :TestExit

for %%j in (store timestamp) do (
  if "%1" == "%%j" (call :TestRunClient %1)
)

for %%j in (frontend stepper) do (
  if "%1" == "%%j" (call :TestRunService %1)
)

shift
goto :TestParseLoop


rem Protective EXIT
rem 
goto :TestExit


:TestRunClient %1
call :TestRun %1 %GOPATH%\src\github.com\Jim3Things\CloudChamber\internal\clients
goto :EOF


:TestRunService %1
call :TestRun %1 %GOPATH%\src\github.com\Jim3Things\CloudChamber\internal\services
goto :EOF



:TestRun %1 %2
pushd %2\%1
go test -v
popd
goto :EOF



:TestHelp

echo.
echo Test ^<test-subject^>
echo.
echo provides for the running of a single set of tests
echo only for one of the following subjects
echo.
echo   frontend
echo   stepper
echo   store
echo   timestamp
echo.
goto :TestExit


:TestExit

endlocal


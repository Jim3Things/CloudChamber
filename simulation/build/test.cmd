@rem
@rem
@rem T E S T . C M D
@rem
@rem Allows one individually selected test, or all, to be run, one at a time
@rem

@if /i "%DbgScript%" == "" @echo off

setlocal

if "%1" == "" goto :TestHelp
if "%1" == "*" goto :TestAll


for %%j in (store timestamp) do (
  if "%1" == "%%j" (call :TestRunClient %1 %2)
)

for %%j in (frontend time) do (
  if "%1" == "%%j" (call :TestRunService %1 %2)
)

for %%j in (deferrable) do (
    if "%1" == "%%j" (call :TestRunTracing common %2)
)

goto :TestExit

:TestAll
for %%j in (timestamp store) do (
    call :TestRunClient %%j %2
)

for %%j in (frontend time) do (
    call :TestRunService %%j %2
)

for %%j in (deferrable) do (
    call :TestRunTracing common %2
)

goto :TestExit

rem Protective EXIT
rem
goto :TestExit


:TestRunClient %1 %2
call :TestRun %1 %GOPATH%\src\github.com\Jim3Things\CloudChamber\internal\clients %2
goto :EOF


:TestRunService %1 %2
call :TestRun %1 %GOPATH%\src\github.com\Jim3Things\CloudChamber\internal\services %2
goto :EOF

:TestRunTracing %1 %2
call :TestRun %1 %GOPATH%\src\github.com\Jim3Things\CloudChamber\internal\tracing\exporters %2
goto :EOF



:TestRun %1 %2 %3
pushd %2\%1
go test %3
popd
goto :EOF



:TestHelp

echo.
echo Test ^<test-subject^> ^<option^>
echo.
echo provides for the running of a single set of tests
echo only for one of the following subjects
echo.
echo   frontend
echo   stepper
echo   store
echo   timestamp
echo   deferrable
echo.
echo The option argument is used when running the targeted test.
echo It defaults to nothing, which runs the tests in quiet more.
echo Setting the option to -v enables verbose tracing output.
goto :TestExit


:TestExit

endlocal


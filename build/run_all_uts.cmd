pushd .

cd /d %GOPATH%\src\github.com\Jim3Things\CloudChamber\internal\clients
go test -v

cd /d %GOPATH%\src\github.com\Jim3Things\CloudChamber\internal\services\stepper
go test -v

popd


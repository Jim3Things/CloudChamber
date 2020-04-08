pushd .

cd /d %GOPATH%\src\github.com\Jim3Things\CloudChamber\internal\clients\timestamp
go test -v

cd /d %GOPATH%\src\github.com\Jim3Things\CloudChamber\internal\clients\store
go test -v

cd /d %GOPATH%\src\github.com\Jim3Things\CloudChamber\internal\services\stepper
go test -v

cd /d %GOPATH%\src\github.com\Jim3Things\CloudChamber\internal\services\frontend
go test -v

popd


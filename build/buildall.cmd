pushd .

cd /d %GOPATH%\src\github.com\Jim3Things\CloudChamber\cmd\controllerd
go build -o %GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments\controllerd.exe main.go

cd /d %GOPATH%\src\github.com\Jim3Things\CloudChamber\cmd\inventoryd
go build -o %GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments\inventoryd.exe main.go

cd /d %GOPATH%\src\github.com\Jim3Things\CloudChamber\cmd\sim_supportd
go build -o %GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments\sim_supportd.exe main.go

cd /d %GOPATH%\src\github.com\Jim3Things\CloudChamber\cmd\web_server
go build -o %GOPATH%\src\github.com\Jim3Things\CloudChamber\deployments\web_server.exe main.go

popd


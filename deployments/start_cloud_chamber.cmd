pushd %gopath%\src\github.com\Jim3Things\CloudChamber\deployments

start controllerd.exe -config=.
start inventoryd.exe -config=.
start sim_supportd.exe -config=.
start web_server.exe -config=.

popd
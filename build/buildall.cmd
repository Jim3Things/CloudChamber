@rem
@rem
@rem B U I L D A L L . C M D
@rem
@rem

@rem @if /i "%DbgScript%" == "" @echo off

rem Generate a usable timestamp that we can use to generate a readme file containing
rem the build time along with anything else we believe to be useful.
rem
rem Note - expects source date format to be US style MONTH/DAY/YEAR
rem 
setlocal

set UpdateDate=%date%
set UpdateTime=%time%

set UpdateYear=%UpdateDate:~10,4%
set UpdateDay=%UpdateDate:~7,2%
set UpdateMonth=%UpdateDate:~4,2%

set UpdateHour=%UpdateTime:~0,2%
set UpdateMinute=%UpdateTime:~3,2%
set UpdateSecond=%UpdateTime:~6,2%

rem Allow for some variants dumping the time var with a leading space rather than a leading zero.
rem
if " " == "%UpdateHour:~0,1%" set UpdateHour=0%UpdateHour:~1,1%

set UpdateDateTime=%UpdateYear%%UpdateMonth%%UpdateDay%-%UpdateHour%%UpdateMinute%%UpdateSecond%


pushd %gopath%\src

set CCRoot=github.com\Jim3Things\CloudChamber
set CCDeployments=%CCRoot%\deployments

protoc --go_out=. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\admin\users.proto
protoc --go_out=. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\common\capacity.proto
protoc --go_out=. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\common\timestamp.proto
protoc --go_out=. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\log\entry.proto
protoc --go_out=. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\inventory\actual.proto
protoc --go_out=. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\inventory\external.proto
protoc --go_out=. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\inventory\internal.proto
protoc --go_out=. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\inventory\target.proto
protoc --go_out=. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\workload\actual.proto
protoc --go_out=. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\workload\external.proto
protoc --go_out=. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\workload\internal.proto
protoc --go_out=. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\workload\target.proto

protoc --go_out=plugins=grpc:. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\monitor\monitor.proto
protoc --go_out=plugins=grpc:. --validate_out=lang=go:. github.com\Jim3Things\CloudChamber\pkg\protos\Stepper\stepper.proto

go build -o github.com\Jim3Things\CloudChamber\deployments\controllerd.exe github.com\Jim3Things\CloudChamber\cmd\controllerd\main.go
go build -o github.com\Jim3Things\CloudChamber\deployments\inventoryd.exe github.com\Jim3Things\CloudChamber\cmd\inventoryd\main.go
go build -o github.com\Jim3Things\CloudChamber\deployments\sim_supportd.exe github.com\Jim3Things\CloudChamber\cmd\sim_supportd\main.go
go build -o github.com\Jim3Things\CloudChamber\deployments\web_server.exe github.com\Jim3Things\CloudChamber\cmd\web_server\main.go

copy github.com\Jim3Things\CloudChamber\Configs\cloudchamber.yaml github.com\Jim3Things\CloudChamber\deployments\cloudchamber.yaml
copy github.com\Jim3Things\CloudChamber\scripts\start_cloud_chamber.cmd github.com\Jim3Things\CloudChamber\deployments\start_cloud_chamber.cmd
copy github.com\Jim3Things\CloudChamber\scripts\startetcd.cmd github.com\Jim3Things\CloudChamber\deployments\startetcd.cmd

echo rem > %CCDeployments%\readme.md
echo rem R E A D M E . m d >> %CCDeployments%\readme.md
echo rem >> %CCDeployments%\readme.md
echo rem >> %CCDeployments%\readme.md
echo BuildTimeStamp %UpdateDateTime% >> %CCDeployments%\readme.md

endlocal
popd


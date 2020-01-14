pushd %gopath%\src
dir $gopath

protoc --go_out=. github.com\Jim3Things\CloudChamber\pkg\protos\common\common.proto
protoc --go_out=plugins=grpc:. github.com\Jim3Things\CloudChamber\pkg\protos\Stepper\stepper.proto
protoc --go_out=. github.com\Jim3Things\CloudChamber\pkg\protos\log\entry.proto
protoc --go_out=. github.com\Jim3Things\CloudChamber\pkg\protos\inventory\actual.proto
protoc --go_out=. github.com\Jim3Things\CloudChamber\pkg\protos\inventory\external.proto
protoc --go_out=. github.com\Jim3Things\CloudChamber\pkg\protos\inventory\internal.proto
protoc --go_out=. github.com\Jim3Things\CloudChamber\pkg\protos\inventory\target.proto
protoc --go_out=plugins=grpc:. github.com\Jim3Things\CloudChamber\pkg\protos\monitor\monitor.proto
protoc --go_out=. github.com\Jim3Things\CloudChamber\pkg\protos\workload\actual.proto
protoc --go_out=. github.com\Jim3Things\CloudChamber\pkg\protos\workload\external.proto
protoc --go_out=. github.com\Jim3Things\CloudChamber\pkg\protos\workload\internal.proto
protoc --go_out=. github.com\Jim3Things\CloudChamber\pkg\protos\workload\target.proto

popd
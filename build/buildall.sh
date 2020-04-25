#~!/bin/bash

export PATH=$PATH:$GOPATH/bin

pushd $GOPATH/src

protoc --go_out=. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/admin/users.proto
protoc --go_out=. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/common/capacity.proto
protoc --go_out=. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/common/timestamp.proto
protoc --go_out=. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/log/entry.proto
protoc --go_out=. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/inventory/actual.proto
protoc --go_out=. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/inventory/external.proto
protoc --go_out=. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/inventory/internal.proto
protoc --go_out=. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/inventory/target.proto
protoc --go_out=. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/workload/actual.proto
protoc --go_out=. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/workload/external.proto
protoc --go_out=. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/workload/internal.proto
protoc --go_out=. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/workload/target.proto

protoc --go_out=plugins=grpc:. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/monitor/monitor.proto
protoc --go_out=plugins=grpc:. --validate_out=lang=go:. github.com/Jim3Things/CloudChamber/pkg/protos/Stepper/stepper.proto

go build -o github.com/Jim3Things/CloudChamber/deployments/controllerd.exe github.com/Jim3Things/CloudChamber/cmd/controllerd/main.go
go build -o github.com/Jim3Things/CloudChamber/deployments/inventoryd.exe github.com/Jim3Things/CloudChamber/cmd/inventoryd/main.go
go build -o github.com/Jim3Things/CloudChamber/deployments/sim_supportd.exe github.com/Jim3Things/CloudChamber/cmd/sim_supportd/main.go
go build -o github.com/Jim3Things/CloudChamber/deployments/web_server.exe github.com/Jim3Things/CloudChamber/cmd/web_server/main.go

cp github.com/Jim3Things/CloudChamber/configs/cloudchamber.yaml github.com/Jim3Things/CloudChamber/deployments/cloudchamber.yaml
cp github.com/Jim3Things/CloudChamber/scripts/start_cloud_chamber.cmd github.com/Jim3Things/CloudChamber/deployments/start_cloud_chamber.cmd
cp github.com/Jim3Things/CloudChamber/scripts/startetcd.cmd github.com/Jim3Things/CloudChamber/deployments/startetcd.cmd

popd

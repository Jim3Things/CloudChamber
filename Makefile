PROJECT = $(GOPATH)/src/github.com/Jim3Things/CloudChamber

PROTO_FILES = \
	pkg/protos/admin/users.proto \
	pkg/protos/common/capacity.proto \
	pkg/protos/common/completion.proto \
	pkg/protos/common/timestamp.proto \
    pkg/protos/log/entry.proto \
    pkg/protos/inventory/actual.proto \
    pkg/protos/inventory/external.proto \
    pkg/protos/inventory/internal.proto \
    pkg/protos/inventory/target.proto \
    pkg/protos/workload/actual.proto \
    pkg/protos/workload/external.proto \
    pkg/protos/workload/internal.proto \
    pkg/protos/workload/target.proto \
    pkg/protos/monitor/monitor.proto \
    pkg/protos/Stepper/stepper.proto


PROTO_GEN_FILES = \
	pkg/protos/admin/users.pb.go\
	pkg/protos/common/capacity.pb.go \
	pkg/protos/common/completion.pb.go \
	pkg/protos/common/timestamp.pb.go \
    pkg/protos/log/entry.pb.go \
    pkg/protos/inventory/actual.pb.go \
    pkg/protos/inventory/external.pb.go \
    pkg/protos/inventory/internal.pb.go \
    pkg/protos/inventory/target.pb.go \
    pkg/protos/workload/actual.pb.go \
    pkg/protos/workload/external.pb.go \
    pkg/protos/workload/internal.pb.go \
    pkg/protos/workload/target.pb.go \
    pkg/protos/monitor/monitor.pb.go \
    pkg/protos/Stepper/stepper.pb.go

SERVICES = \
	deployments/controllerd.exe \
	deployments/inventoryd.exe \
	deployments/sim_supportd.exe \
	deployments/web_server.exe

VERSION_MARKER = \
	pkg/version/generated.go \
	pkg/version/version_stamp.md

ARTIFACTS = \
	deployments/readme.md \
	deployments/cloudchamber.yaml \
	deployments/start_cloud_chamber.cmd \
	deployments/startetcd.cmd

PROTOC = protoc --go_out=$(GOPATH)/src --proto_path=. --proto_path=$(GOPATH)/src
GRPC_PROTOC = protoc --go_out=plugins=grpc:$(GOPATH)/src --proto_path=. --proto_path=$(GOPATH)/src

CP = cp

pkg/protos/admin/users.pb.go : pkg/protos/admin/users.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/common/capacity.pb.go : pkg/protos/common/capacity.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/common/completion.pb.go : pkg/protos/common/completion.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/common/timestamp.pb.go : pkg/protos/common/timestamp.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/log/entry.pb.go: pkg/protos/log/entry.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/inventory/actual.pb.go: pkg/protos/inventory/actual.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/inventory/external.pb.go: pkg/protos/inventory/external.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/inventory/internal.pb.go: pkg/protos/inventory/internal.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/inventory/target.pb.go: pkg/protos/inventory/target.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/workload/actual.pb.go: pkg/protos/workload/actual.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/workload/external.pb.go: pkg/protos/workload/external.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/workload/internal.pb.go: pkg/protos/workload/internal.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/workload/target.pb.go: pkg/protos/workload/target.proto
	$(PROTOC) $(PROJECT)/$<

pkg/protos/monitor/monitor.pb.go: pkg/protos/monitor/monitor.proto
	$(GRPC_PROTOC) $(PROJECT)/$<

pkg/protos/Stepper/stepper.pb.go: pkg/protos/Stepper/stepper.proto
	$(GRPC_PROTOC) $(PROJECT)/$<


deployments/controllerd.exe: cmd/controllerd/main.go
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

deployments/inventoryd.exe: cmd/inventoryd/main.go
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

deployments/sim_supportd.exe: cmd/sim_supportd/main.go
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

deployments/web_server.exe: cmd/web_server/main.go
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

deployments/readme.md: pkg/version/version_stamp.md
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/cloudchamber.yaml: Configs/cloudchamber.yaml
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/start_cloud_chamber.cmd: scripts/start_cloud_chamber.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/startetcd.cmd: scripts/startetcd.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

all: \
	clean \
	build \
	run_tests

build: \
	protogen \
	version \
	service_build \
	copy_to

protogen: $(PROTO_GEN_FILES)

version:
	go generate $(PROJECT)/pkg/version/version.go

service_build: $(SERVICES)

copy_to: $(ARTIFACTS)

run_tests:
	go test $(PROJECT)/internal/clients/store
	go test $(PROJECT)/internal/clients/timestamp
	go test $(PROJECT)/internal/services/frontend
	go test $(PROJECT)/internal/services/stepper_actor

clean:
	$(RM) deployments/*
	$(RM) $(PROTO_GEN_FILES)
	$(RM) $(VERSION_MARKER)

test: \
	protogen \
	version \
	run_tests
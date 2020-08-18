PROJECT = $(GOPATH)/src/github.com/Jim3Things/CloudChamber

PROJECT_UI = ../cloud_chamber_react_ts/build

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
    pkg/protos/Stepper/stepper.proto \
    pkg/protos/trace_sink/trace_sink.proto


PROTO_GEN_FILES = \
    pkg/protos/admin/users.pb.go \
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
    pkg/protos/Stepper/stepper.pb.go \
    pkg/protos/trace_sink/trace_sink.pb.go


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
    deployments/StartAll.cmd \
    deployments/StartCloudChamber.cmd \
    deployments/StartEtcd.cmd \
    deployments/MonitorEtcd.cmd



INSTALL_KIT = $(SERVICES) $(ARTIFACTS)



ifdef SYSTEMDRIVE

INSTALL_TARGET = $(SYSTEMDRIVE)/CloudChamber/Files

else

INSTALL_TARGET = ~/CloudChamber/Files

endif



PROTOC_BASE = protoc --proto_path=. --proto_path=$(GOPATH)/src

PROTOC_PBUF = $(PROTOC_BASE) --go_out=$(GOPATH)/src 
PROTOC_GRPC = $(PROTOC_BASE) --go_out=plugins=grpc:$(GOPATH)/src 


CP = cp
CP-RECURSIVE = $(CP) -r

MD = mkdir -p



all: build run_tests

build: $(SERVICES) $(ARTIFACTS)

protogen: $(PROTO_GEN_FILES)

version: $(VERSION_MARKER)

service_build: $(SERVICES)

copy_to: $(ARTIFACTS)


.PHONY : install

install: $(INSTALL_KIT)
	$(MD) $(INSTALL_TARGET)
	$(CP) $(INSTALL_KIT) $(INSTALL_TARGET)/
	$(CP) $(PROJECT_UI)/*.* $(INSTALL_TARGET)/
	$(CP-RECURSIVE) $(PROJECT_UI)/static $(INSTALL_TARGET)/


.PHONY : run_tests

run_tests: $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go test $(PROJECT)/internal/clients/store
	go test $(PROJECT)/internal/clients/timestamp
	go test $(PROJECT)/internal/services/frontend
	go test $(PROJECT)/internal/services/stepper_actor
	go test $(PROJECT)/internal/tracing/exporters/common


.PHONY : clean

clean:
	$(RM) $(SERVICES) $(ARTIFACTS) $(PROTO_GEN_FILES) $(VERSION_MARKER)


.PHONY : test

test: run_tests


%.pb.go : %.proto
	$(PROTOC_PBUF) $(PROJECT)/$<


pkg/protos/monitor/monitor.pb.go: pkg/protos/monitor/monitor.proto
	$(PROTOC_GRPC) $(PROJECT)/$<

pkg/protos/Stepper/stepper.pb.go: pkg/protos/Stepper/stepper.proto
	$(PROTOC_GRPC) $(PROJECT)/$<

pkg/protos/trace_sink/trace_sink.pb.go: pkg/protos/trace_sink/trace_sink.proto
	$(PROTOC_GRPC) $(PROJECT)/$<



$(VERSION_MARKER) &: pkg/version/version.go
	go generate $(PROJECT)/$<

deployments/controllerd.exe: cmd/controllerd/main.go   $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

deployments/inventoryd.exe: cmd/inventoryd/main.go     $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

deployments/sim_supportd.exe: cmd/sim_supportd/main.go $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

deployments/web_server.exe: cmd/web_server/main.go     $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

deployments/readme.md: pkg/version/version_stamp.md
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/cloudchamber.yaml: configs/cloudchamber.yaml
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/StartEtcd.cmd: scripts/StartEtcd.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/StartAll.cmd : scripts/StartAll.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/StartCloudChamber.cmd : scripts/StartCloudChamber.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/MonitorEtcd.cmd : scripts/MonitorEtcd.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

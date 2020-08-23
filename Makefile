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


SRC_CONFIG = \
	internal/config/settings.go

SRC_FRONTEND = \
	$(SRC_CONFIG) \
	$(SRC_STORE) \
	$(SRC_TIMESTAMP) \
	$(SRC_TRACING) \
	$(SRC_TRACING_CLIENT) \
	$(SRC_TRACING_SERVER) \
	internal/services/frontend/DBInventory.go \
	internal/services/frontend/DBUsers.go \
	internal/services/frontend/errors.go \
	internal/services/frontend/frontend.go \
	internal/services/frontend/inventory.go \
	internal/services/frontend/ping.go \
	internal/services/frontend/session_manager.go \
	internal/services/frontend/stepper.go \
	internal/services/frontend/users.go \
	internal/services/frontend/workloads.go

SRC_MONITOR = \
	internal/services/monitor/monitor.go

SRC_SM = \
	$(SRC_TRACING_SERVER) \
	internal/sm/sm.go

SRC_STEPPER_ACTOR = \
	$(SRC_SM) \
	$(SRC_TRACING) \
	$(SRC_TRACING_SERVER) \
	internal/services/stepper_actor/actor.go \
	internal/services/stepper_actor/adapter.go \
	internal/services/stepper_actor/sm.go

SRC_STORE = \
	$(SRC_TRACING) \
	$(SRC_TRACING_SERVER) \
	internal/clients/store/store.go \
	internal/clients/store/storeapi.go \
	internal/clients/store/errors.go

SRC_TIMESTAMP = \
	internal/clients/timestamp/timestamp.go

SRC_TRACINGSINK = \
	internal/services/tracing_sink/sink.go

SRC_TRACING = \
	internal/tracing/constants.go \
	internal/tracing/StackData.go

SRC_TRACING_EXPORTERS_COMMON = \
	$(SRC_TRACING) \
	internal/tracing/exporters/common/common.go \
	internal/tracing/exporters/common/deferrable.go

SRC_TRACING_EXPORTERS_IO_WRITER = \
	$(SRC_TRACING_EXPORTERS_COMMON) \
	internal/tracing/exporters/io_writer/exporter.go

SRC_TRACING_EXPORTERS_PRODUCTION = \
	$(SRC_TRACING_EXPORTERS_COMMON) \
	internal/tracing/exporters/production/exporter.go

SRC_TRACING_EXPORTERS = \
	$(SRC_TRACING_EXPORTERS_IO_WRITER) \
	$(SRC_TRACING_EXPORTERS_PRODUCTION) \
	internal/tracing/exporters/exporters.go \
	internal/tracing/exporters/init.go 

SRC_TRACING_SETUP = \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_EXPORTERS_IO_WRITER) \
	$(SRC_TRACING_EXPORTERS_PRODUCTION) \
	internal/tracing/setup/config.go \

SRC_TRACING_SERVER = \
	$(SRC_TRACING) \
	internal/tracing/server/actor_interceptor.go \
	internal/tracing/server/span_map.go \
	internal/tracing/server/tracing.go

SRC_TRACING_CLIENT = \
	internal/tracing/client/tracing.go \

SRC_VERSION = \
	pkg/version/version.go


SRC_CONTROLLER = \
	cmd/controllerd/main.go \
	$(SRC_CONFIG) \
	$(SRC_MONITOR) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SERVER) \
	$(SRC_TRACING_SETUP)

SRC_INVENTORY = \
	cmd/inventoryd/main.go \
	$(SRC_CONFIG) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SETUP)

SRC_SIMSUPPORT = \
	cmd/sim_supportd/main.go \
	$(SRC_CONFIG) \
	$(SRC_STEPPER_ACTOR) \
	$(SRC_TRACINGSINK) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SERVER) \
	$(SRC_TRACING_SETUP)

SRC_WEBSERVER = \
	cmd/web_server/main.go \
	$(SRC_CONFIG) \
	$(SRC_FRONTEND) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SETUP)


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
    deployments/Deploy.cmd \
    deployments/StartAll.cmd \
    deployments/StartCloudChamber.cmd \
    deployments/StartEtcd.cmd \
    deployments/MonitorEtcd.cmd



INSTALL_KIT = $(SERVICES) $(ARTIFACTS)



ifdef CC_INSTALL_TARGET

INSTALL_TARGET = $(CC_INSTALL_TARGET)/Files

else ifdef SYSTEMDRIVE

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

RM-RECURSIVE = $(RM) -r



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


.PHONY : install_clean

install_clean:

	$(RM-RECURSIVE) $(INSTALL_TARGET)/static
	$(RM-RECURSIVE) $(INSTALL_TARGET)



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



$(VERSION_MARKER) &: $(SRC_VERSION)
	go generate $(PROJECT)/$<

deployments/controllerd.exe:  $(SRC_CONTROLLER) $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

deployments/inventoryd.exe:   $(SRC_INVENTORY)  $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

deployments/sim_supportd.exe: $(SRC_SIMSUPPORT) $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

deployments/web_server.exe:   $(SRC_WEBSERVER)  $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

deployments/readme.md: pkg/version/version_stamp.md
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/cloudchamber.yaml: configs/cloudchamber.yaml
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/Deploy.cmd: scripts/Deploy.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/StartEtcd.cmd: scripts/StartEtcd.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/StartAll.cmd : scripts/StartAll.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/StartCloudChamber.cmd : scripts/StartCloudChamber.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

deployments/MonitorEtcd.cmd : scripts/MonitorEtcd.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

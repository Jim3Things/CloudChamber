PROJECT = $(GOPATH)/src/github.com/Jim3Things/CloudChamber

PROJECT_UI = clients/observer/build

PROTO_FILES = \
    pkg/protos/admin/simulation.proto \
    pkg/protos/admin/users.proto \
    pkg/protos/common/completion.proto \
    pkg/protos/common/timestamp.proto \
    pkg/protos/log/entry.proto \
    pkg/protos/inventory/capacity.proto \
    pkg/protos/inventory/common.proto \
    pkg/protos/inventory/actual.proto \
    pkg/protos/inventory/definition.proto \
    pkg/protos/inventory/external.proto \
    pkg/protos/inventory/internal.proto \
    pkg/protos/inventory/store.proto \
    pkg/protos/inventory/target.proto \
    pkg/protos/workload/actual.proto \
    pkg/protos/workload/external.proto \
    pkg/protos/workload/internal.proto \
    pkg/protos/workload/target.proto \
    pkg/protos/services/inventory.proto \
    pkg/protos/services/monitor.proto \
    pkg/protos/services/requests.proto \
    pkg/protos/services/stepper.proto \
    pkg/protos/services/trace_sink.proto


PROTO_GEN_FILES = \
    pkg/protos/admin/simulation.pb.go \
    pkg/protos/admin/users.pb.go \
    pkg/protos/common/completion.pb.go \
    pkg/protos/common/timestamp.pb.go \
    pkg/protos/log/entry.pb.go \
    pkg/protos/inventory/capacity.pb.go \
    pkg/protos/inventory/common.pb.go \
    pkg/protos/inventory/actual.pb.go \
    pkg/protos/inventory/definition.pb.go \
    pkg/protos/inventory/external.pb.go \
    pkg/protos/inventory/internal.pb.go \
    pkg/protos/inventory/store.pb.go \
    pkg/protos/inventory/target.pb.go \
    pkg/protos/workload/actual.pb.go \
    pkg/protos/workload/external.pb.go \
    pkg/protos/workload/internal.pb.go \
    pkg/protos/workload/target.pb.go \
    pkg/protos/services/inventory.pb.go \
    pkg/protos/services/monitor.pb.go \
    pkg/protos/services/requests.pb.go \
    pkg/protos/services/stepper.pb.go \
    pkg/protos/services/trace_sink.pb.go


ProdFiles = $(filter-out %_test.go, $(wildcard $(1)/*.go))

SRC_ERRORS = \
	$(call ProdFiles, pkg/errors)

SRC_CONFIG = \
	$(SRC_ERRORS) \
	$(call ProdFiles, internal/config)

SRC_FRONTEND = \
	$(SRC_CONFIG) \
	$(SRC_ERRORS) \
	$(SRC_INVENTORY_CLIENT) \
	$(SRC_STORE) \
	$(SRC_TIMESTAMP) \
	$(SRC_TRACING) \
	$(SRC_TRACING_CLIENT) \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, internal/services/frontend)

SRC_MONITOR = \
	$(call ProdFiles, internal/services/monitor)

SRC_INVENTORY_SERVICE = \
	$(SRC_TIMESTAMP) \
	$(SRC_TRACING) \
	$(SRC_TRACING_CLIENT) \
	$(call ProdFiles, internal/services/inventory)

SRC_SM = \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, internal/sm)

SRC_STEPPER_ACTOR = \
    $(SRC_SM) \
	$(SRC_TRACING) \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, internal/services/stepper)

SRC_INVENTORY_CLIENT = \
	$(SRC_TRACING) \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, internal/clients/inventory)

SRC_STORE = \
	$(SRC_ERRORS) \
	$(SRC_TRACING) \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, internal/clients/store)

SRC_TIMESTAMP = \
	$(SRC_ERRORS) \
	$(call ProdFiles, internal/clients/timestamp)

SRC_TRACING = \
	$(call ProdFiles, internal/tracing)

SRC_TRACING_EXPORTERS = \
	$(SRC_TRACING) \
	$(call ProdFiles, internal/tracing/exporters)

SRC_TRACING_SETUP = \
	$(SRC_TRACING_EXPORTERS) \
	$(call ProdFiles, internal/tracing/setup)

SRC_TRACING_SERVER = \
	$(SRC_TRACING) \
	$(call ProdFiles, internal/tracing/server)

SRC_TRACING_CLIENT = \
	$(SRC_ERRORS) \
	$(call ProdFiles, internal/tracing/client)

SRC_VERSION = \
	pkg/version/version.go


SRC_CONTROLLER = \
	$(call ProdFiles, cmd/controllerd) \
	$(SRC_CONFIG) \
	$(SRC_MONITOR) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SERVER) \
	$(SRC_TRACING_SETUP)

SRC_INVENTORY = \
	$(call ProdFiles, cmd/inventoryd) \
	$(SRC_CONFIG) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SETUP) \
	$(SRC_INVENTORY_SERVICE)

SRC_SIMSUPPORT = \
	$(call ProdFiles, cmd/sim_supportd) \
	$(SRC_CONFIG) \
	$(SRC_STEPPER_ACTOR) \
	$(SRC_TRACINGSINK) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SERVER) \
	$(SRC_TRACING_SETUP)

SRC_WEBSERVER = \
	$(call ProdFiles, cmd/web_server) \
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
    deployments/inventory.yaml \
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



PROTOC_BASE = protoc --proto_path=$(GOPATH)/src

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
	go test -count=1 $(PROJECT)/internal/clients/inventory
	go test -count=1 $(PROJECT)/internal/clients/store
	go test -count=1 $(PROJECT)/internal/clients/timestamp
	go test -count=1 $(PROJECT)/internal/config
	go test -count=1 $(PROJECT)/internal/services/frontend
	go test -count=1 $(PROJECT)/internal/services/inventory
	go test -count=1 $(PROJECT)/internal/services/repair_manager/inventory
	go test -count=1 $(PROJECT)/internal/services/repair_manager/ruler
	go test -count=1 $(PROJECT)/internal/services/stepper
	go test -count=1 $(PROJECT)/internal/services/tracing_sink
	go test -count=1 $(PROJECT)/internal/tracing/exporters


.PHONY : clean

clean:
	$(RM) $(SERVICES) $(ARTIFACTS) $(PROTO_GEN_FILES) $(VERSION_MARKER)


.PHONY : test

test: run_tests


%.pb.go : %.proto
	$(PROTOC_PBUF) $(PROJECT)/$<


pkg/protos/services/inventory.pb.go: pkg/protos/services/inventory.proto
	$(PROTOC_GRPC) $(PROJECT)/$<

pkg/protos/services/monitor.pb.go: pkg/protos/services/monitor.proto
	$(PROTOC_GRPC) $(PROJECT)/$<

pkg/protos/services/stepper.pb.go: pkg/protos/services/stepper.proto
	$(PROTOC_GRPC) $(PROJECT)/$<

pkg/protos/services/trace_sink.pb.go: pkg/protos/services/trace_sink.proto
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

deployments/inventory.yaml: configs/inventory.yaml
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

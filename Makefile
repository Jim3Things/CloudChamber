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

PROTO_CS_GEN_FILES = \
    pkg/protos/admin/simulation.pb.cs \
    pkg/protos/admin/users.pb.cs \
    pkg/protos/common/completion.pb.cs \
    pkg/protos/common/timestamp.pb.cs \
    pkg/protos/log/entry.pb.cs \
    pkg/protos/inventory/capacity.pb.cs \
    pkg/protos/inventory/common.pb.cs \
    pkg/protos/inventory/actual.pb.cs \
    pkg/protos/inventory/definition.pb.cs \
    pkg/protos/inventory/external.pb.cs \
    pkg/protos/inventory/internal.pb.cs \
    pkg/protos/inventory/store.pb.cs \
    pkg/protos/inventory/target.pb.cs \
    pkg/protos/workload/actual.pb.cs \
    pkg/protos/workload/external.pb.cs \
    pkg/protos/workload/internal.pb.cs \
    pkg/protos/workload/target.pb.cs \
    pkg/protos/services/requests.pb.cs

ProdFiles = $(filter-out %_test.go, $(wildcard $(1)/*.go))

SRC_ERRORS = \
	$(call ProdFiles, simulation/pkg/errors)

SRC_CONFIG = \
	$(SRC_ERRORS) \
	$(call ProdFiles, simulation/internal/config)

SRC_FRONTEND = \
	$(SRC_CONFIG) \
	$(SRC_ERRORS) \
	$(SRC_INVENTORY_CLIENT) \
	$(SRC_STORE) \
	$(SRC_TIMESTAMP) \
	$(SRC_TRACING) \
	$(SRC_TRACING_CLIENT) \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, simulation/internal/services/frontend)

SRC_MONITOR = \
	$(call ProdFiles, simulation/internal/services/monitor)

SRC_INVENTORY_SERVICE = \
	$(SRC_TIMESTAMP) \
	$(SRC_TRACING) \
	$(SRC_TRACING_CLIENT) \
	$(call ProdFiles, simulation/internal/services/inventory)

SRC_SM = \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, internal/sm)

SRC_STEPPER_ACTOR = \
    $(SRC_SM) \
	$(SRC_TRACING) \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, simulation/internal/services/stepper)

SRC_INVENTORY_CLIENT = \
	$(SRC_TRACING) \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, simulation/internal/clients/inventory)

SRC_STORE = \
	$(SRC_ERRORS) \
	$(SRC_TRACING) \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, simulation/internal/clients/store)

SRC_TIMESTAMP = \
	$(SRC_ERRORS) \
	$(call ProdFiles, simulation/internal/clients/timestamp)

SRC_TRACING = \
	$(call ProdFiles, simulation/internal/tracing)

SRC_TRACING_EXPORTERS = \
	$(SRC_TRACING) \
	$(call ProdFiles, simulation/internal/tracing/exporters)

SRC_TRACING_SETUP = \
	$(SRC_TRACING_EXPORTERS) \
	$(call ProdFiles, simulation/internal/tracing/setup)

SRC_TRACING_SERVER = \
	$(SRC_TRACING) \
	$(call ProdFiles, simulation/internal/tracing/server)

SRC_TRACING_CLIENT = \
	$(SRC_ERRORS) \
	$(call ProdFiles, simulation/internal/tracing/client)

SRC_VERSION = \
	simulation/pkg/version/version.go


SRC_CONTROLLER = \
	$(call ProdFiles, simulation/cmd/controllerd) \
	$(SRC_CONFIG) \
	$(SRC_MONITOR) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SERVER) \
	$(SRC_TRACING_SETUP)

SRC_INVENTORY = \
	$(call ProdFiles, simulation/cmd/inventoryd) \
	$(SRC_CONFIG) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SETUP) \
	$(SRC_INVENTORY_SERVICE)

SRC_SIMSUPPORT = \
	$(call ProdFiles, simulation/cmd/sim_supportd) \
	$(SRC_CONFIG) \
	$(SRC_STEPPER_ACTOR) \
	$(SRC_TRACINGSINK) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SERVER) \
	$(SRC_TRACING_SETUP)

SRC_WEBSERVER = \
	$(call ProdFiles, simulation/cmd/web_server) \
	$(SRC_CONFIG) \
	$(SRC_FRONTEND) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SETUP)


SERVICES = \
    simulation/deployments/controllerd.exe \
    simulation/deployments/inventoryd.exe \
    simulation/deployments/sim_supportd.exe \
    simulation/deployments/web_server.exe

VERSION_MARKER = \
    simulation/pkg/version/generated.go \
    simulation/pkg/version/version_stamp.md

ARTIFACTS = \
    simulation/deployments/readme.md \
    simulation/deployments/cloudchamber.yaml \
    simulation/deployments/inventory.yaml \
    simulation/deployments/Deploy.cmd \
    simulation/deployments/StartAll.cmd \
    simulation/deployments/StartCloudChamber.cmd \
    simulation/deployments/StartEtcd.cmd \
    simulation/deployments/MonitorEtcd.cmd



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
PROTOC_PBUF_CS = $(PROTOC_BASE) --csharp_out=$(PROJECT)/pkg/protos --csharp_opt=file_extension=.pb.cs,base_namespace=CloudChamber.Protos
PROTOC_GRPC = $(PROTOC_BASE) --go_out=plugins=grpc:$(GOPATH)/src


CP = cp
CP-RECURSIVE = $(CP) -r

MD = mkdir -p

RM-RECURSIVE = $(RM) -r



all: build run_tests

build: $(PROTO_CS_GEN_FILES) $(SERVICES) $(ARTIFACTS)

protogen: $(PROTO_GEN_FILES) $(PROTO_GEN_CS_FILES)

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
	go test -count=1 $(PROJECT)/simulation/internal/clients/inventory
	go test -count=1 $(PROJECT)/simulation/internal/clients/store
	go test -count=1 $(PROJECT)/simulation/internal/clients/timestamp
	go test -count=1 $(PROJECT)/simulation/internal/common
	go test -count=1 $(PROJECT)/simulation/internal/services/frontend
	go test -count=1 $(PROJECT)/simulation/internal/services/inventory
	go test -count=1 $(PROJECT)/simulation/internal/services/repair_manager/inventory
	go test -count=1 $(PROJECT)/simulation/internal/services/repair_manager/ruler
	go test -count=1 $(PROJECT)/simulation/internal/services/stepper
	go test -count=1 $(PROJECT)/simulation/internal/services/tracing_sink
	go test -count=1 $(PROJECT)/simulation/internal/tracing/exporters


.PHONY : clean

clean:
	$(RM) $(SERVICES) $(ARTIFACTS) $(PROTO_GEN_FILES) $(PROTO_CS_GEN_FILES) $(VERSION_MARKER)


.PHONY : test

test: run_tests


%.pb.go : %.proto
	$(PROTOC_PBUF) $(PROJECT)/$<

%.pb.cs : %.proto
	$(PROTOC_PBUF_CS) $(PROJECT)/$<

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

simulation/deployments/controllerd.exe:  $(SRC_CONTROLLER) $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

simulation/deployments/inventoryd.exe:   $(SRC_INVENTORY)  $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

simulation/deployments/sim_supportd.exe: $(SRC_SIMSUPPORT) $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

simulation/deployments/web_server.exe:   $(SRC_WEBSERVER)  $(PROTO_GEN_FILES) $(VERSION_MARKER)
	go build -o $(PROJECT)/$@ $(PROJECT)/$<

simulation/deployments/readme.md: simulation/pkg/version/version_stamp.md
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

simulation/deployments/cloudchamber.yaml: simulation/configs/cloudchamber.yaml
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

simulation/deployments/inventory.yaml: simulation/configs/inventory.yaml
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

simulation/deployments/Deploy.cmd: simulation/scripts/Deploy.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

simulation/deployments/StartEtcd.cmd: simulation/scripts/StartEtcd.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

simulation/deployments/StartAll.cmd : simulation/scripts/StartAll.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

simulation/deployments/StartCloudChamber.cmd : simulation/scripts/StartCloudChamber.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

simulation/deployments/MonitorEtcd.cmd : simulation/scripts/MonitorEtcd.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

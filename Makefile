PROJECT = $(GOPATH)/src/github.com/Jim3Things/CloudChamber

OBSERVER_UI = clients/observer
OBSERVER_UI_BUILD = $(OBSERVER_UI)/build

KIT_BUILD = simulation/deployments

ProdFiles = $(filter-out %_test.go, $(wildcard $(1)/*.go))
CssFiles = $(wildcard $(1)/*.css)
SvgFiles = $(wildcard $(1)/*.svg)
TsFiles  = $(wildcard $(1)/*.ts)
TsxFiles = $(wildcard $(1)/*.tsx)


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
    pkg/protos/workload/observed.proto \
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
    pkg/protos/inventory/observed.pb.go \
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
    pkg/protos/inventory/observed.pb.cs \
    pkg/protos/inventory/store.pb.cs \
    pkg/protos/inventory/target.pb.cs \
    pkg/protos/workload/actual.pb.cs \
    pkg/protos/workload/external.pb.cs \
    pkg/protos/workload/internal.pb.cs \
    pkg/protos/workload/target.pb.cs \
    pkg/protos/services/requests.pb.cs

PROTO_TS_GEN_FILES = \
    pkg/protos/admin/simulation.ts \
    pkg/protos/admin/users.ts \
    pkg/protos/common/completion.ts \
    pkg/protos/common/timestamp.ts \
    pkg/protos/inventory/capacity.ts \
    pkg/protos/inventory/common.ts \
    pkg/protos/inventory/actual.ts \
    pkg/protos/inventory/definition.ts \
    pkg/protos/inventory/external.ts \
    pkg/protos/inventory/internal.ts \
    pkg/protos/inventory/observed.ts \
    pkg/protos/inventory/store.ts \
    pkg/protos/inventory/target.ts \
    pkg/protos/log/entry.ts \
    pkg/protos/workload/actual.ts \
    pkg/protos/workload/external.ts \
    pkg/protos/workload/internal.ts \
    pkg/protos/workload/target.ts \
    pkg/protos/services/requests.ts

PROTO_TSX_FILES = \
    $(call TsxFiles, pkg/protos/admin) \
    $(call TsxFiles, pkg/protos/common) \
    $(call TsxFiles, pkg/protos/inventory) \
    $(call TsxFiles, pkg/protos/log) \
    $(call TsxFiles, pkg/protos/services) \
    $(call TsxFiles, pkg/protos)


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

SRC_ARTIFACTS = \
	simulation/pkg/version/readme.md \
	simulation/configs/cloudchamber.yaml \
	simulation/configs/inventory.yaml \
	simulation/scripts/Deploy.cmd \
	simulation/scripts/StartEtcd.cmd \
	simulation/scripts/StartAll.cmd \
	simulation/scripts/StartCloudChamber.cmd \
	simulation/scripts/MonitorEtcd.cmd


ARTIFACTS = $(addprefix $(KIT_BUILD)/, $(notdir $(SRC_ARTIFACTS)))

OBSERVER = \
    $(OBSERVER_UI_BUILD)/asset-manifest.json \
    $(OBSERVER_UI_BUILD)/index.html

SERVICES = \
    $(KIT_BUILD)/controllerd.exe \
    $(KIT_BUILD)/inventoryd.exe \
    $(KIT_BUILD)/sim_supportd.exe \
    $(KIT_BUILD)/web_server.exe

VERSION_MARKER = \
    simulation/pkg/version/generated.go \
    simulation/pkg/version/version_stamp.md \
	simulation/pkg/version/readme.md


INSTALL_KIT = $(SERVICES) $(ARTIFACTS) $(OBSERVER)


ifdef CC_INSTALL_TARGET

INSTALL_TARGET = $(CC_INSTALL_TARGET)/Files

else ifdef SYSTEMDRIVE

INSTALL_TARGET = $(SYSTEMDRIVE)/CloudChamber/Files

else

INSTALL_TARGET = ~/CloudChamber/Files

endif


PROTOC_BASE = protoc --proto_path=$(GOPATH)/src

PROTOC_PBUF    = $(PROTOC_BASE) --go_out=$(GOPATH)/src
PROTOC_PBUF_CS = $(PROTOC_BASE) --csharp_out=$(PROJECT)/pkg/protos --csharp_opt=file_extension=.pb.cs,base_namespace=CloudChamber.Protos
PROTOC_PBUF_TS = $(PROTOC_BASE) --ts_proto_out=$(GOPATH)/src --ts_proto_opt=outputEncodeMethods=false,outputPartialMethods=false,outputClientImpl=false
PROTOC_GRPC    = $(PROTOC_BASE) --go_out=plugins=grpc:$(GOPATH)/src


CP = cp
CP-RECURSIVE = $(CP) -r

MD = mkdir -p

RM-RECURSIVE = $(RM) -r

TOUCH = touch --no-create


define run-proto-grpc =
	$(PROTOC_GRPC) $(PROJECT)/$^
endef

define run-go-build =
	go build -o $(PROJECT)/$@ $(PROJECT)/$<
endef


all: build run_tests

build: $(SERVICES) $(ARTIFACTS) $(OBSERVER_UI)

protogen: $(PROTO_GEN_FILES) $(PROTO_GEN_CS_FILES) $(PROTO_TS_GEN_FILES)

version: $(VERSION_MARKER)

service_build: $(SERVICES)

copy_to: $(ARTIFACTS)


.PHONY : install

install: $(INSTALL_KIT)
	$(MD) $(INSTALL_TARGET)
	$(CP) $(INSTALL_KIT) $(INSTALL_TARGET)/
	$(CP) $(OBSERVER_UI_BUILD)/*.* $(INSTALL_TARGET)/
	$(CP-RECURSIVE) $(OBSERVER_UI_BUILD)/static $(INSTALL_TARGET)/


.PHONY : install_clean

install_clean:
	$(RM-RECURSIVE) $(INSTALL_TARGET)/static
	$(RM-RECURSIVE) $(INSTALL_TARGET)


.PHONY : kit

kit : $(INSTALL_KIT)
	$(CP) $(OBSERVER_UI_BUILD)/*.* $(KIT_BUILD)/
	$(CP-RECURSIVE) $(OBSERVER_UI_BUILD)/static $(KIT_BUILD)/
	$(TOUCH) $(KIT_BUILD)


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
	$(RM) $(SERVICES) $(ARTIFACTS) $(PROTO_GEN_FILES) $(PROTO_CS_GEN_FILES) $(PROTO_TS_GEN_FILES) $(VERSION_MARKER)
	$(MAKE) -C $(OBSERVER_UI) clean


.PHONY : test

test: run_tests


.PHONY : ui

ui:
	$(MAKE) -C $(OBSERVER_UI) build


%.pb.go : %.proto
	$(PROTOC_PBUF) $(PROJECT)/$<

%.pb.cs : %.proto
	$(PROTOC_PBUF_CS) $(PROJECT)/$<

%.ts : %.proto
	$(PROTOC_PBUF_TS) $(PROJECT)/$<

%.ts_ref : %.ts
	echo ******* Check for out of date $@ *******
%.tsx : %.ts_ref
	echo ******* Check for out of date $@ *******


$(ARTIFACTS) &: $(SRC_ARTIFACTS)
	$(CP) $(SRC_ARTIFACTS) $(dir $@)

$(VERSION_MARKER) &: $(SRC_VERSION)
	go generate $(PROJECT)/$<
	$(CP) simulation/pkg/version/version_stamp.md simulation/pkg/version/readme.md


pkg/protos/services/inventory.pb.go: pkg/protos/services/inventory.proto
	$(run-proto-grpc)

pkg/protos/services/monitor.pb.go: pkg/protos/services/monitor.proto
	$(run-proto-grpc)

pkg/protos/services/stepper.pb.go: pkg/protos/services/stepper.proto
	$(run-proto-grpc)

pkg/protos/services/trace_sink.pb.go: pkg/protos/services/trace_sink.proto
	$(run-proto-grpc)


simulation/deployments/controllerd.exe:  $(SRC_CONTROLLER) $(PROTO_GEN_FILES) $(VERSION_MARKER)
	$(run-go-build)

simulation/deployments/inventoryd.exe:   $(SRC_INVENTORY)  $(PROTO_GEN_FILES) $(VERSION_MARKER)
	$(run-go-build)

simulation/deployments/sim_supportd.exe: $(SRC_SIMSUPPORT) $(PROTO_GEN_FILES) $(VERSION_MARKER)
	$(run-go-build)

simulation/deployments/web_server.exe:   $(SRC_WEBSERVER)  $(PROTO_GEN_FILES) $(VERSION_MARKER)
	$(run-go-build)

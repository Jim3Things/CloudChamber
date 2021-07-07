export PROJECT = $(GOPATH)/src/github.com/Jim3Things/CloudChamber

OBSERVER_UI = clients/observer
OBSERVER_UI_BUILD = clients/observer/build

KIT_BUILD = simulation/deployments

ProdFiles = $(filter-out %_test.go, $(wildcard $(1)/*.go))

CssFiles = $(wildcard $(1)/*.css)
SvgFiles = $(wildcard $(1)/*.svg)
TsFiles  = $(wildcard $(1)/*.ts)
TsxFiles = $(wildcard $(1)/*.tsx)

PROTO_FILES = \
	simulation/pkg/protos/admin/simulation.proto \
	simulation/pkg/protos/admin/users.proto \
	simulation/pkg/protos/common/completion.proto \
	simulation/pkg/protos/common/timestamp.proto \
	simulation/pkg/protos/log/entry.proto \
	simulation/pkg/protos/inventory/capacity.proto \
	simulation/pkg/protos/inventory/common.proto \
	simulation/pkg/protos/inventory/actual.proto \
	simulation/pkg/protos/inventory/definition.proto \
	simulation/pkg/protos/inventory/external.proto \
	simulation/pkg/protos/inventory/internal.proto \
	simulation/pkg/protos/inventory/observed.proto \
	simulation/pkg/protos/inventory/store.proto \
	simulation/pkg/protos/inventory/target.proto \
	simulation/pkg/protos/workload/actual.proto \
	simulation/pkg/protos/workload/external.proto \
	simulation/pkg/protos/workload/internal.proto \
	simulation/pkg/protos/workload/target.proto \
	simulation/pkg/protos/services/inventory.proto \
	simulation/pkg/protos/services/monitor.proto \
	simulation/pkg/protos/services/requests.proto \
	simulation/pkg/protos/services/stepper.proto \
	simulation/pkg/protos/services/traceSink.proto

# Every proto file is compiled into go, csharp, and typescript results.
PROTO_GEN_FILES = $(subst .proto,.pb.go,$(PROTO_FILES))
PROTO_CS_GEN_FILES = $(subst .proto,.pb.cs,$(PROTO_FILES))
PROTO_TS_GEN_FILES = $(subst .proto,.ts,$(PROTO_FILES))

# Construct the list of files that will be exported from the simulator to
# use in other subprojects.
EXPORT_CS_GEN_FILES = $(subst simulation/pkg/,pkg/,$(PROTO_CS_GEN_FILES))
export EXPORT_TSX_FILES = \
	pkg/protos/utils.tsx \
	pkg/protos/admin/users.tsx \
	pkg/protos/common/Timestamp.tsx \
	pkg/protos/inventory/capacity.tsx \
	pkg/protos/inventory/common.tsx \
	pkg/protos/inventory/external.tsx \
	pkg/protos/log/entry.tsx \
	pkg/protos/services/requests.tsx

SRC_COMMON = \
	$(call ProdFiles, simulation/internal/common)

SRC_ERRORS = \
	$(call ProdFiles, simulation/pkg/errors)

SRC_CONFIG = \
	$(SRC_COMMON) \
	$(SRC_ERRORS) \
	$(call ProdFiles, simulation/internal/config)

SRC_LIMITS_CLIENT = \
	$(call ProdFiles, simulation/internal/clients/limits)

SRC_NAMESPACE_CLIENT = \
	$(SRC_LIMITS_CLIENT) \
	$(call ProdFiles, simulation/internal/clients/namespace)

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

SRC_INVENTORY_CLIENT = \
	$(SRC_COMMON) \
	$(SRC_NAMESPACE_CLIENT) \
	$(SRC_TRACING) \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, simulation/internal/clients/inventory)

SRC_FRONTEND = \
	$(SRC_COMMON) \
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

SRC_REPAIR_MANAGER = \
	$(SRC_ERRORS) \
	$(SRC_COMMON) \
	$(SRC_TRACING) \
	$(call ProdFiles, simulation/internal/services/rules_manager) \
	$(call ProdFiles, simulation/internal/services/rules_manager/inventory) \
	$(call ProdFiles, simulation/internal/services/rules_manager/inventory/rules) \
	$(call ProdFiles, simulation/internal/services/rules_manager/ruler) \
	$(call ProdFiles, simulation/internal/services/rules_manager/workload)

SRC_SM = \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, internal/sm)

SRC_STEPPER = \
	$(SRC_COMMON) \
    $(SRC_SM) \
	$(SRC_TRACING) \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, simulation/internal/services/stepper)

SRC_STORE = \
	$(SRC_ERRORS) \
	$(SRC_TRACING) \
	$(SRC_TRACING_SERVER) \
	$(call ProdFiles, simulation/internal/clients/store)

SRC_TIMESTAMP = \
	$(SRC_COMMON) \
	$(SRC_ERRORS) \
	$(call ProdFiles, simulation/internal/clients/timestamp)

SRC_INVENTORY_SERVICE = \
	$(SRC_COMMON) \
	$(SRC_CONFIG) \
	$(SRC_ERRORS) \
	$(SRC_INVENTORY_CLIENT) \
	$(SRC_STORE) \
	$(SRC_TIMESTAMP) \
	$(SRC_TRACING) \
	$(SRC_TRACING_CLIENT) \
	$(call ProdFiles, simulation/internal/services/inventory) \
	$(call ProdFiles, simulation/internal/services/inventory/messages)

SRC_VERSION = \
	simulation/pkg/version/version.go


SRC_CONTROLLER = \
	$(call ProdFiles, simulation/cmd/controllerd) \
	$(SRC_REPAIR_MANAGER) \
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
	$(SRC_STEPPER) \
	$(SRC_TRACINGSINK) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SERVER) \
	$(SRC_TRACING_SETUP)

SRC_WEBSERVER = \
	$(call ProdFiles, simulation/cmd/web_server) \
	$(SRC_CONFIG) \
	$(SRC_FRONTEND) \
	$(SRC_TRACING_EXPORTERS) \
	$(SRC_TRACING_SETUP) \
	$(SRC_LIMITS_CLIENT)

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

# Define how to build the Go protobuf backends, which can either include GRPC
# support or not.
PROTOC_GO = $(PROTOC_BASE) --go_out=$(GOPATH)/src
PROTOC_GRPC = $(PROTOC_BASE) --go_out=plugins=grpc:$(GOPATH)/src

# Define how to build the typescript backends.  Note that this output is not
# used in the build.  It is used to verify that the hand-crafted .tsx files
# continue to reflect the protobuf definitions.  This is done by comparing the
# generated .ts files with the .ts_ref files, which would have been saved as
# part of the commit that edited the .tsx file.
PROTOC_TS = $(PROTOC_BASE) --ts_proto_out=$(GOPATH)/src --ts_proto_opt=outputEncodeMethods=false,outputPartialMethods=false,outputClientImpl=false

# Define how to build the csharp backends.  The complexity here stems from the
# different expectations on casing - csharp expects the namespaces to be proper
# cased, but the folders are lower-cased.  PROTOC_CS_GEN is a function that
# allows this difference.
PROTOC_CS_BASE = $(PROTOC_BASE) --csharp_out=$(PROJECT)/simulation/pkg/protos/$$dir$$ --csharp_opt=file_extension=.pb.cs,base_namespace=CloudChamber.Protos.$$NS$$
PROTOC_CS_GEN = $(subst $$dir$$,$(1),$(subst $$NS$$,$(2),$(PROTOC_CS_BASE)))

UT_GO = go test -count=1
BUILD_GO = go build -o

CP = cp
CP-RECURSIVE = $(CP) -r

MD = mkdir -p
RM-RECURSIVE = $(RM) -r

TOUCH = touch --no-create

all: build run_tests

build: $(EXPORT_CS_GEN_FILES) $(EXPORT_TSX_FILES) $(SERVICES) $(ARTIFACTS) $(OBSERVER)

protogen: $(PROTO_GEN_FILES) $(PROTO_CS_GEN_FILES) $(PROTO_TS_GEN_FILES)

version: $(VERSION_MARKER)

service_build: $(SERVICES)

copy_to: $(ARTIFACTS)


.PHONY : install

install: $(INSTALL_KIT)
	$(MD) $(INSTALL_TARGET)/static
	$(RM-RECURSIVE) $(INSTALL_TARGET)/static
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
	$(UT_GO) $(PROJECT)/simulation/internal/clients/inventory
	$(UT_GO) $(PROJECT)/simulation/internal/clients/store
	$(UT_GO) $(PROJECT)/simulation/internal/clients/timestamp
	$(UT_GO) $(PROJECT)/simulation/internal/common
	$(UT_GO) $(PROJECT)/simulation/internal/services/frontend
	$(UT_GO) $(PROJECT)/simulation/internal/services/inventory
	$(UT_GO) $(PROJECT)/simulation/internal/services/repair_manager/inventory
	$(UT_GO) $(PROJECT)/simulation/internal/services/repair_manager/ruler
	$(UT_GO) $(PROJECT)/simulation/internal/services/stepper
	$(UT_GO) $(PROJECT)/simulation/internal/services/tracing_sink
	$(UT_GO) $(PROJECT)/simulation/internal/tracing/exporters
	$(MAKE) -C $(OBSERVER_UI) test


.PHONY : clean

clean:
	$(RM) $(SERVICES) $(ARTIFACTS) $(PROTO_GEN_FILES) $(PROTO_CS_GEN_FILES) $(PROTO_TS_GEN_FILES) $(EXPORT_CS_GEN_FILES) $(EXPORT_TSX_FILES) $(VERSION_MARKER)
	$(MAKE) -C $(OBSERVER_UI) clean
	$(RM-RECURSIVE) $(KIT_BUILD)


.PHONY : test

test: run_tests

.PHONY : exported

exported: $(EXPORT_CS_GEN_FILES) $(EXPORT_TSX_FILES)

.PHONY : ui

ui:
	$(MAKE) -C $(OBSERVER_UI) build


$(OBSERVER) &: $(EXPORT_CS_GEN_FILES) $(OBSERVER_UI)
	$(MAKE) -C $(OBSERVER_UI) build

%.pb.go : %.proto
	$(PROTOC_GO) $(PROJECT)/$<

simulation/pkg/protos/admin/%.pb.cs : simulation/pkg/protos/admin/%.proto
	$(call PROTOC_CS_GEN,admin,Admin) $(PROJECT)/$<

simulation/pkg/protos/common/%.pb.cs : simulation/pkg/protos/common/%.proto
	$(call PROTOC_CS_GEN,common,Common) $(PROJECT)/$<

simulation/pkg/protos/log/%.pb.cs : simulation/pkg/protos/log/%.proto
	$(call PROTOC_CS_GEN,log,Log) $(PROJECT)/$<

simulation/pkg/protos/inventory/%.pb.cs : simulation/pkg/protos/inventory/%.proto
	$(call PROTOC_CS_GEN,inventory,Inventory) $(PROJECT)/$<

simulation/pkg/protos/workload/%.pb.cs : simulation/pkg/protos/workload/%.proto
	$(call PROTOC_CS_GEN,workload,Workload) $(PROJECT)/$<

simulation/pkg/protos/services/%.pb.cs : simulation/pkg/protos/services/%.proto
	$(call PROTOC_CS_GEN,services,Services) $(PROJECT)/$<

%.ts : %.proto
	$(PROTOC_TS) $(PROJECT)/$<

simulation/pkg/protos/services/%.pb.go : simulation/pkg/protos/services/%.proto
	$(PROTOC_GRPC) $(PROJECT)/$<

pkg/protos/%.pb.cs : simulation/pkg/protos/%.pb.cs
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

pkg/protos/%.tsx : simulation/pkg/protos/%.tsx
	$(CP) $(PROJECT)/$< $(PROJECT)/$@



$(SERVICES) $(ARTIFACTS) $(OBSERVER) : | $(KIT_BUILD)

$(KIT_BUILD) :
	$(MD) $(KIT_BUILD)


$(VERSION_MARKER) &: $(SRC_VERSION)
	go generate $(PROJECT)/$<
	$(CP) simulation/pkg/version/version_stamp.md simulation/pkg/version/readme.md

simulation/deployments/controllerd.exe:  $(SRC_CONTROLLER) $(PROTO_GEN_FILES) $(VERSION_MARKER)
	$(BUILD_GO) $(PROJECT)/$@ $(PROJECT)/$<

simulation/deployments/inventoryd.exe:   $(SRC_INVENTORY)  $(PROTO_GEN_FILES) $(VERSION_MARKER)
	$(BUILD_GO) $(PROJECT)/$@ $(PROJECT)/$<

simulation/deployments/sim_supportd.exe: $(SRC_SIMSUPPORT) $(PROTO_GEN_FILES) $(VERSION_MARKER)
	$(BUILD_GO) $(PROJECT)/$@ $(PROJECT)/$<

simulation/deployments/web_server.exe:   $(SRC_WEBSERVER)  $(PROTO_GEN_FILES) $(VERSION_MARKER)
	$(BUILD_GO) $(PROJECT)/$@ $(PROJECT)/$<

simulation/deployments/readme.md: simulation/pkg/version/readme.md
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

simulation/deployments/cloudchamber.yaml: simulation/configs/cloudchamber.yaml
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

simulation/deployments/inventory.yaml: simulation/configs/Inventory.yaml
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

simulation/deployments/%.cmd: simulation/scripts/%.cmd
	$(CP) $(PROJECT)/$< $(PROJECT)/$@

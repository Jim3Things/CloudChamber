PROJECT  = $(GOPATH)/src/github.com/Jim3Things/CloudChamber
PROTODIR = $(PROJECT)/pkg/protos

ProdFiles  = $(filter-out %_test.go, $(wildcard $(1)/*.go))
CssFiles   = $(wildcard $(1)/*.css)
SvgFiles   = $(wildcard $(1)/*.svg)
TsFiles    = $(wildcard $(1)/*.ts)
TsxFiles   = $(wildcard $(1)/*.tsx)
ProtoFiles = $(wildcard $(1)/*.proto)


PROTO_FILES = \
    admin/simulation.proto \
    admin/users.proto \
    common/completion.proto \
    common/timestamp.proto \
    log/entry.proto \
    inventory/capacity.proto \
    inventory/common.proto \
    inventory/actual.proto \
    inventory/definition.proto \
    inventory/external.proto \
    inventory/internal.proto \
    inventory/store.proto \
    inventory/target.proto \
    workload/actual.proto \
    workload/external.proto \
    workload/internal.proto \
    workload/target.proto \
    services/inventory.proto \
    services/monitor.proto \
    services/requests.proto \
    services/stepper.proto \
    services/trace_sink.proto

PROTO_PB_GEN_FILES = \
    admin/simulation.pb.go \
    admin/users.pb.go \
    common/completion.pb.go \
    common/timestamp.pb.go \
    log/entry.pb.go \
    inventory/capacity.pb.go \
    inventory/common.pb.go \
    inventory/actual.pb.go \
    inventory/definition.pb.go \
    inventory/external.pb.go \
    inventory/internal.pb.go \
    inventory/store.pb.go \
    inventory/target.pb.go \
    workload/actual.pb.go \
    workload/external.pb.go \
    workload/internal.pb.go \
    workload/target.pb.go \
    services/inventory.pb.go \
    services/monitor.pb.go \
    services/requests.pb.go \
    services/stepper.pb.go \
    services/trace_sink.pb.go

PROTO_CS_GEN_FILES = \
    admin/simulation.pb.cs \
    admin/users.pb.cs \
    common/completion.pb.cs \
    common/timestamp.pb.cs \
    log/entry.pb.cs \
    inventory/capacity.pb.cs \
    inventory/common.pb.cs \
    inventory/actual.pb.cs \
    inventory/definition.pb.cs \
    inventory/external.pb.cs \
    inventory/internal.pb.cs \
    inventory/store.pb.cs \
    inventory/target.pb.cs \
    workload/actual.pb.cs \
    workload/external.pb.cs \
    workload/internal.pb.cs \
    workload/target.pb.cs \
    services/requests.pb.cs

PROTO_TS_GEN_FILES = \
    admin/simulation.ts \
    admin/users.ts \
    common/completion.ts \
    common/timestamp.ts \
    inventory/capacity.ts \
    inventory/common.ts \
    inventory/actual.ts \
    inventory/definition.ts \
    inventory/external.ts \
    inventory/internal.ts \
    inventory/store.ts \
    inventory/target.ts \
    log/entry.ts \
    workload/actual.ts \
    workload/external.ts \
    workload/internal.ts \
    workload/target.ts \
    services/requests.ts

PROTO_TSX_FILES = \
    $(call TsxFiles, admin) \
    $(call TsxFiles, common) \
    $(call TsxFiles, inventory) \
    $(call TsxFiles, log) \
    $(call TsxFiles, services) \
    $(call TsxFiles, .)




PROTOC_BASE = protoc --proto_path=$(GOPATH)/src

PROTOC_GRPC    = $(PROTOC_BASE) --go_out=plugins=grpc:$(GOPATH)/src
PROTOC_PBUF    = $(PROTOC_BASE) --go_out=$(GOPATH)/src

#PROTOC_PBUF_CS = $(PROTOC_BASE) --csharp_out=$(PROTODIR) --csharp_opt=file_extension=.pb.cs,base_namespace=CloudChamber.Protos
#PROTOC_PBUF_TS = $(PROTOC_BASE) --ts_proto_out=$(GOPATH)/src --ts_proto_opt=outputEncodeMethods=false,outputPartialMethods=false,outputClientImpl=false


CP = cp
CP-RECURSIVE = $(CP) -r

MD = mkdir -p

RM-RECURSIVE = $(RM) -r


define run-proto-grpc = 
	$(PROTOC_GRPC) $(PROTODIR)/$^
endef






all: build

build: $(PROTO_PB_GEN_FILES) $(PROTO_GEN_CS_FILES) $(PROTO_TS_GEN_FILES)


.PHONY : clean

clean:
	$(RM) $(PROTO_PB_GEN_FILES) $(PROTO_CS_GEN_FILES) $(PROTO_TS_GEN_FILES)




%.pb.go : %.proto
	$(PROTOC_PBUF) $(PROTODIR)/$<

%.pb.cs : %.proto
	$(PROTOC_PBUF_CS) $(PROTODIR)/$<

%.ts : %.proto
	$(PROTOC_PBUF_TS) $(PROTODIR)/$<

%.ts_ref : %.ts 
	echo ******* Check for out of date $@ *******

%.tsx : %.ts_ref
	echo ******* Check for out of date $@ *******


#services/inventory.pb.go: services/inventory.proto
#	$(PROTOC_GRPC) $(PROTODIR)/$<

#services/monitor.pb.go: services/monitor.proto
#	$(PROTOC_GRPC) $(PROTODIR)/$<

#services/stepper.pb.go: services/stepper.proto
#	$(PROTOC_GRPC) $(PROTODIR)/$<

#services/trace_sink.pb.go: services/trace_sink.proto
#	$(PROTOC_GRPC) $(PROTODIR)/$<


services/inventory.pb.go: services/inventory.proto
	$(run-proto-grpc)

services/monitor.pb.go: services/monitor.proto
	$(run-proto-grpc)

services/stepper.pb.go: services/stepper.proto
	$(run-proto-grpc)

services/trace_sink.pb.go: services/trace_sink.proto
	$(run-proto-grpc)


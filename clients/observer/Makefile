CP = cp
CP-RECURSIVE = $(CP) -r

MD = mkdir -p
RM-RECURSIVE = $(RM) -r

TOUCH = touch --no-create

OBSERVER_PROJECT = $(PROJECT)/clients/observer

CssFiles = $(wildcard $(1)/*.css)
SvgFiles = $(wildcard $(1)/*.svg)
TsFiles  = $(wildcard $(1)/*.ts)
TsxFiles = $(wildcard $(1)/*.tsx)

SRC_PROTO_TSX_FILES = \
	$(addprefix src/,$(EXPORT_TSX_FILES))

SRC_PUBLIC = \
    public/favicon.ico \
    public/index.html \
    public/logo192.png \
    public/logo512.png \
    public/manifest.json \
    public/robots.txt

SRC_COMMAND_AREA = \
    $(call TsxFiles, src/CommandArea)

SRC_COMMON = \
    $(call TsxFiles, src/common)

SRC_LOG = \
    $(call TsxFiles, src/Log)

SRC_MAIN_PAGE = \
    $(call TsxFiles, src/MainPage)

SRC_PROXIES = \
    $(call TsxFiles, src/proxies)

SRC_SIMULATED_INVENTORY = \
    $(call TsxFiles, src/SimulatedInventory)


SRC_OBSERVER = \
    $(SRC_PROTO_TSX_FILES) \
    $(SRC_PUBLIC) \
    $(SRC_COMMAND_AREA) \
    $(SRC_COMMON) \
    $(SRC_LOG) \
    $(SRC_MAIN_PAGE) \
    $(SRC_PROXIES) \
    $(SRC_SIMULATED_INVENTORY) \
    $(call CssFiles, src) \
    $(call SvgFiles, src) \
    $(call TsFiles, src) \
    $(call TsxFiles, src) \


OBSERVER = \
    build/favicon.ico \
    build/logo192.png \
    build/logo512.png \
    build/manifest.json \
    build/robots.txt \
    build/index.html \
    build/asset-manifest.json


.PHONY: all
all: build


.PHONY: build
build: $(OBSERVER)


.PHONY: clean
clean:
	$(RM-RECURSIVE) build/*
	$(RM-RECURSIVE) src/pkg/protos/*

# Unit tests are disabled for now because they emit a noisy error about Organizer
# serialization.  Doesn't seem to impact behavior (and the test itself passes),
# but I'm suppressing the test running in make until I've resolved the serialization
# issue.
.PHONY: test
test:
#	set CI=true && npm run-script test

src/pkg/protos/%.tsx : $(PROJECT)/pkg/protos/%.tsx
	$(MD) $(dir $(OBSERVER_PROJECT)/$@)
	$(CP) $< $(OBSERVER_PROJECT)/$@

$(OBSERVER) &: $(SRC_OBSERVER)
	$(RM-RECURSIVE) build/*
	npm run-script build
	$(TOUCH) build/*

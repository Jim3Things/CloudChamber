
RM-RECURSIVE = $(RM) -r

TOUCH = touch --no-create


CssFiles = $(wildcard $(1)/*.css)
SvgFiles = $(wildcard $(1)/*.svg)
TsFiles  = $(wildcard $(1)/*.ts)
TsxFiles = $(wildcard $(1)/*.tsx)


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
    $(PROTO_TSX_FILES) \
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
    build//asset-manifest.json


.PHONY: all
all: build


.PHONY: build 
build: $(OBSERVER)


.PHONY: clean
clean:
	$(RM-RECURSIVE) build/*


$(OBSERVER) &: $(SRC_OBSERVER)
	$(RM-RECURSIVE) build/*
	npm run-script build
	$(TOUCH) build/*




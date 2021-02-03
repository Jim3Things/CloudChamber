PROJECT = $(GOPATH)/src/github.com/Jim3Things/CloudChamber

PROJECT_UI = ../cloud_chamber_react_ts

GENERATED = $(PROJECT_UI)/build


RM-RECURSIVE = $(RM) -r



all: build


build:
	npm run-script build


clean:
	$(RM-RECURSIVE) $(GENERATED)


rebuild: clean build


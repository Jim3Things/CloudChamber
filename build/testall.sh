#!/bin/bash

pushd .

cd $GOPATH/src/github.com/Jim3Things/CloudChamber/internal/clients/timestamp
go test -v

cd $GOPATH/src/github.com/Jim3Things/CloudChamber/internal/clients/store
go test -v

cd $GOPATH/src/github.com/Jim3Things/CloudChamber/internal/services/stepper_actor
go test -v

cd $GOPATH/src/github.com/Jim3Things/CloudChamber/internal/services/frontend
go test -v

popd


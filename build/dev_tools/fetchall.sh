#!/bin/bash

pushd $GOPATH/src

go get github.com/golang/protobuf
go get github.com/golang/protobuf/protoc-gen-go
go get github.com/envoyproxy/protoc-gen-validate

go get github.com/gorilla/mux
go get github.com/gorilla/securecookie
go get github.com/gorilla/sessions

#go get github.com/davecgh/go-spew/spew
#go get github.com/pmezard/go-difflib/difflib
go get github.com/spf13/viper
go get github.com/stretchr/testify/assert

go get go.etcd.io/etcd
go get go.etcd.io/etcd/etdctl

go get go.opentelemetry.io/otel

go get google.golang.org/grpc

go get golang.org/x/crypto/...

go get github.com/Jim3Things/CloudChamber

popd

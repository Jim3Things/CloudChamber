#!/bin/bash

pushd $GOPATH/src

go get -u github.com/golang/protobuf
go get -u github.com/golang/protobuf/protoc-gen-go
go get -u github.com/envoyproxy/protoc-gen-validate

go get -u github.com/gorilla/mux
go get -u github.com/gorilla/securecookie
go get -u github.com/gorilla/sessions

#go get -u github.com/davecgh/go-spew/spew
#go get -u github.com/pmezard/go-difflib/difflib
go get -u github.com/spf13/viper
go get -u github.com/stretchr/testify/assert

go get -u go.etcd.io/etcd
go get -u go.etcd.io/etcd/etdctl

go get -u go.opentelemetry.io/otel

go get -u google.golang.org/grpc

go get -u golang.org/x/crypto/...

popd

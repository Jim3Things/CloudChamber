if /i "%GOPATH%" == "" (
  echo Need to set GOPATH environment variable first before running any go commands
  goto :EOF
)

pushd %GOPATH%\src

go get github.com/golang/protobuf


rem fetch & install the protobuf compiler for Go
rem
go get github.com/golang/protobuf/protoc-gen-go
pushd %GOPATH%\src\github.com\golang\protobuf\protoc-gen-go
go install .
popd


rem fetch & install the protobuf validation plugin
rem
go get -d github.com/envoyproxy/protoc-gen-validate
pushd %GOPATH%\src\github.com\envoyproxy\protoc-gen-validate
go install .
popd


go get github.com/gorilla/mux
go get github.com/gorilla/securecookie
go get github.com/gorilla/sessions

go get github.com/davecgh/go-spew/spew
go get github.com/pmezard/go-difflib/difflib
go get github.com/spf13/viper
go get github.com/stretchr/testify/assert

go get go.etcd.io/etcd

go get go.opentelemetry.io/otel

go get google.golang.org/grpc

go get golang.org/x/crypto/...

go get github.com/Jim3Things/CloudChamber

popd
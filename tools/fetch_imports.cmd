if /i "%GOPATH%" == "" (
  echo Need to set GOPATH environment variable first before running any go commands
  goto :EOF
)

go get -u go.opentelemetry.io/otel

go get google.golang.org/grpc

go get -u golang.org/x/crypto/...

go get "github.com/etcd-io/etcd"

go get "github.com/golang/protobuf/ptypes/duration"
go get "github.com/golang/protobuf/ptypes/empty"

go get "github.com/gorilla/mux"
go get "github.com/gorilla/securecookie"
go get "github.com/gorilla/sessions"

go get "github.com/stretchr/testify/assert"

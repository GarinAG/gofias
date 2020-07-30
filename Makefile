#!make

run-update:
	cd ./app && go run ./application/cli/main.go --config-path=../ update --skip-houses --skip-clear

run-index:
	cd ./app && go run ./application/cli/main.go --config-path=../ index

run-grpc:
	cd ./app && go run ./application/grpc/main.go --config-path=../

protoc:
	protoc -I. -I$GOPATH/src -I$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.6/third_party/googleapis --go_out=plugins=grpc:. app/interfaces/grpc/proto/v1/*/*.proto && \
	protoc -I/usr/local/include -I. -I$GOPATH/src -I$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.6/third_party/googleapis --grpc-gateway_out=logtostderr=true:.  app/interfaces/grpc/proto/v1/*/*.proto && \
	protoc -I/usr/local/include -I. -I$GOPATH/src -I$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.6/third_party/googleapis --swagger_out=logtostderr=true:.  app/interfaces/grpc/proto/v1/*/*.proto

#!make
include .env
export $(shell sed 's/=.*//' .env)

run-update:
	cd ./app && go run ./application/cli/main.go --config-path=../ update --skip-houses --skip-clear

run-index:
	cd ./app && go run ./application/cli/main.go --config-path=../ index

run-grpc:
	cd ./app && go run ./application/grpc/main.go --config-path=../

protoc:
	export GOOGLEAPIS=$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.6/third_party/googleapis;\
	protoc -I. -I$GOPATH/src -I$GOOGLEAPIS --go_out=plugins=grpc:. app/interfaces/grpc/proto/version/*.proto && \
	protoc -I/usr/local/include -I. -I$GOOGLEAPIS --grpc-gateway_out=logtostderr=true:.  app/interfaces/grpc/proto/version/*.proto && \
	protoc -I/usr/local/include -I. -I$GOOGLEAPIS --swagger_out=logtostderr=true:.  app/interfaces/grpc/proto/version/*.proto;\
	protoc -I. -I$GOPATH/src -I$GOOGLEAPIS --go_out=plugins=grpc:. app/interfaces/grpc/proto/v1/address/*.proto && \
	protoc -I/usr/local/include -I. -I$GOOGLEAPIS --grpc-gateway_out=logtostderr=true:.  app/interfaces/grpc/proto/v1/address/*.proto && \
	protoc -I/usr/local/include -I. -I$GOOGLEAPIS --swagger_out=logtostderr=true:.  app/interfaces/grpc/proto/v1/address/*.proto;\
	protoc -I. -I$GOPATH/src -I$GOOGLEAPIS --go_out=plugins=grpc:. app/interfaces/grpc/proto/v1/health/*.proto && \
	protoc -I/usr/local/include -I. -I$GOOGLEAPIS --grpc-gateway_out=logtostderr=true:.  app/interfaces/grpc/proto/v1/health/*.proto && \
	protoc -I/usr/local/include -I. -I$GOOGLEAPIS --swagger_out=logtostderr=true:.  app/interfaces/grpc/proto/v1/health/*.proto;

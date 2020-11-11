#!make
include .env
export $(shell sed 's/=.*//' .env)
export GOROOT=/usr/local/go
export GOPATH=$(HOME)/go
export GOBIN=$(GOPATH)/bin
export PATH:=$(PATH):$(GOROOT):$(GOPATH):$(GOBIN)
export SWAGGER_OPTIONS=$(GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.15.2
export GOOGLEAPIS=$(SWAGGER_OPTIONS)/third_party/googleapis

test:
	echo $(PATH)

up:
	docker-compose up -d

build:
	docker-compose up -d -build

run-update:
	cd ./app && go run ./application/cli/main.go --config-path=../ update --skip-houses --skip-clear

run-index:
	cd ./app && go run ./application/cli/main.go --config-path=../ index

run-grpc:
	cd ./app && go run ./application/grpc/main.go --config-path=../

protoc:
	protoc -I. -I$(GOPATH)/src -I$(GOOGLEAPIS) -I$(SWAGGER_OPTIONS) --go_out=plugins=grpc:. app/interfaces/grpc/proto/v1/fias/*.proto && \
	protoc -I/usr/local/include -I. -I$(GOOGLEAPIS) -I$(SWAGGER_OPTIONS) --grpc-gateway_out=logtostderr=true:.  app/interfaces/grpc/proto/v1/fias/*.proto && \
	protoc -I/usr/local/include -I. -I$(GOOGLEAPIS) -I$(SWAGGER_OPTIONS) --swagger_out=logtostderr=true:.  app/interfaces/grpc/proto/v1/fias/*.proto;

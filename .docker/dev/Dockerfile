# build stage
FROM golang:alpine as build

ARG GRPC_PORT=50051
ARG GRPC_GATEWAY_PORT=8081

RUN mkdir /app
WORKDIR /app
COPY ./app .

RUN go clean --modcache
RUN go get github.com/githubnemo/CompileDaemon
RUN go mod download
RUN go build /app/application/grpc/main.go

EXPOSE ${GRPC_PORT} ${GRPC_GATEWAY_PORT}
ENTRYPOINT CompileDaemon --build="go build /app/application/grpc/main.go" --command="./main --config-type=env"
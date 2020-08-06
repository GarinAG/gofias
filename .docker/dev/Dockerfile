# build stage
FROM golang:alpine as build

RUN mkdir /app
WORKDIR /app
COPY ./app .

RUN go clean --modcache
RUN go get github.com/githubnemo/CompileDaemon
RUN go mod download
RUN go build /app/application/grpc/main.go

EXPOSE 8080
ENTRYPOINT CompileDaemon --build="go build /app/application/grpc/main.go" --command="./main --config-type=env"
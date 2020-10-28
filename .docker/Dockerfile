# build stage
FROM golang:alpine as build

ARG GRPC_PORT=50051
ARG GRPC_GATEWAY_PORT=8081

RUN mkdir /app
WORKDIR /app
COPY ./app .

RUN go clean --modcache
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main /app/application/grpc/main.go

# final stage
FROM scratch
COPY --from=build /app/main /

EXPOSE ${GRPC_PORT} ${GRPC_GATEWAY_PORT}
ENTRYPOINT ["/main --config-type=env"]
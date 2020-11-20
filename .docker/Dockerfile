# build stage
FROM golang:alpine as builder

ARG GRPC_PORT=50051
ARG GRPC_GATEWAY_PORT=8081

RUN mkdir /app
WORKDIR /app
COPY app/ .

RUN go clean --modcache
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main /app/application/grpc/main.go

# final stage
FROM scratch
WORKDIR /
COPY --from=builder /app/main .
COPY --from=builder /app/swagger ./swagger
COPY --from=builder /app/interfaces/grpc ./interfaces/grpc

EXPOSE ${GRPC_PORT} ${GRPC_GATEWAY_PORT}
ENTRYPOINT ["/main", "--config-type=env"]
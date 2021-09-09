# build stage
FROM golang:alpine as build

RUN mkdir /app
WORKDIR /app
COPY ./app .

RUN go clean --modcache
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main /app/application/grpc/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o fias /app/application/cli/main.go

# final stage
FROM scratch
COPY --from=build /app/main /
EXPOSE 8080
ENTRYPOINT ["/main --config-type=env"]

package grpc

import (
	"context"
	"flag"
	"github.com/GarinAG/gofias/domain/address/service"
	versionService "github.com/GarinAG/gofias/domain/version/service"
	addressHandler "github.com/GarinAG/gofias/infrastructure/persistence/grpc/handler"
	"github.com/GarinAG/gofias/infrastructure/registry"
	"github.com/GarinAG/gofias/interfaces"
	grpcHandlerAddress "github.com/GarinAG/gofias/interfaces/grpc/proto/v1/address"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

type GrpcServer struct {
	Server         *grpc.Server
	Logger         interfaces.LoggerInterface
	AddressService *service.AddressService
	HouseService   *service.HouseImportService
	VersionService *versionService.VersionService
}

func NewGrpcServer(ctn *registry.Container) *GrpcServer {
	logger := ctn.Resolve("logger").(interfaces.LoggerInterface)
	defer func() {
		if r := recover(); r != nil {
			logger.WithFields(interfaces.LoggerFields{"error": r}).Panic("Program fatal error")
			os.Exit(1)
		}
	}()
	gserver := grpc.NewServer()
	grpcHandlerAddress.RegisterAddressHandlerServer(gserver, addressHandler.NewAddressHandler(ctn.Resolve("addressService").(*service.AddressService)))
	reflection.Register(gserver)

	return &GrpcServer{
		Server:         gserver,
		Logger:         logger,
		AddressService: ctn.Resolve("addressService").(*service.AddressService),
		HouseService:   ctn.Resolve("houseService").(*service.HouseImportService),
		VersionService: ctn.Resolve("versionService").(*versionService.VersionService),
	}
}

func (g *GrpcServer) Run() error {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go g.grpcService(&wg)
	go g.proxyService(&wg)
	wg.Wait()
	return nil
}

func (g *GrpcServer) grpcService(wg *sync.WaitGroup) {
	defer wg.Done()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		g.Server.GracefulStop()
	}()

	if err := g.Serve(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (g *GrpcServer) Serve() error {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}

	return g.Server.Serve(listener)
}

func (g *GrpcServer) proxyService(wg *sync.WaitGroup) {
	defer wg.Done()
	var grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:50051", "gRPC server endpoint")
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := grpcHandlerAddress.RegisterAddressHandlerHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		g.Logger.Fatal("error reg endpoint", err)
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	g.Logger.Info("Start Http server on port: 8081")
	g.Logger.Fatal(http.ListenAndServe(":8081", mux).Error())
}

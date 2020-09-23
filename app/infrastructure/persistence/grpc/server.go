package grpc

import (
	"context"
	"github.com/GarinAG/gofias/domain/address/service"
	versionService "github.com/GarinAG/gofias/domain/version/service"
	grpcHandlerHealth "github.com/GarinAG/gofias/infrastructure/persistence/grpc/dto/health"
	grpcHandlerAddressV1 "github.com/GarinAG/gofias/infrastructure/persistence/grpc/dto/v1/address"
	grpcHandlerVersion "github.com/GarinAG/gofias/infrastructure/persistence/grpc/dto/version"
	handlers "github.com/GarinAG/gofias/infrastructure/persistence/grpc/handler"
	"github.com/GarinAG/gofias/infrastructure/registry"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

var Version = "2.0.0"

type GrpcServer struct {
	Server         *grpc.Server
	Logger         interfaces.LoggerInterface
	Config         interfaces.ConfigInterface
	AddressService *service.AddressService
	HouseService   *service.HouseImportService
	VersionService *versionService.VersionService
}

var globalLogger interfaces.LoggerInterface

func NewGrpcServer(ctn *registry.Container) *GrpcServer {
	logger := ctn.Resolve("logger").(interfaces.LoggerInterface)
	globalLogger = logger

	defer func() {
		if r := recover(); r != nil {
			logger.WithFields(interfaces.LoggerFields{"error": r}).Panic("Program fatal error")
			os.Exit(1)
		}
	}()
	server := grpc.NewServer(grpc.UnaryInterceptor(serverInterceptor))
	grpcHandlerAddressV1.RegisterAddressHandlerServer(server,
		handlers.NewAddressHandler(
			ctn.Resolve("addressService").(*service.AddressService),
			ctn.Resolve("houseService").(*service.HouseService),
		))
	grpcHandlerHealth.RegisterHealthHandlerServer(server, handlers.NewHealthHandler())
	grpcHandlerVersion.RegisterVersionHandlerServer(server, handlers.NewVersionHandler(ctn.Resolve("versionService").(*versionService.VersionService), Version))

	reflection.Register(server)

	return &GrpcServer{
		Server:         server,
		Logger:         logger,
		Config:         ctn.Resolve("config").(interfaces.ConfigInterface),
		AddressService: ctn.Resolve("addressService").(*service.AddressService),
		HouseService:   ctn.Resolve("houseImportService").(*service.HouseImportService),
		VersionService: ctn.Resolve("versionService").(*versionService.VersionService),
	}
}

func (g *GrpcServer) Run() error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go g.grpcService(&wg)
	if g.Config.GetBool("grpc.gateway.enable") {
		wg.Add(1)
		go g.proxyService(&wg)
	}
	wg.Wait()
	return nil
}

func (g *GrpcServer) grpcService(wg *sync.WaitGroup) {
	defer wg.Done()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		oscall := <-c
		g.Logger.Info("system call:%+v", oscall)
		g.Server.GracefulStop()
		os.Exit(0)
	}()

	if err := g.Serve(); err != nil {
		g.Logger.Fatal("failed to serve", err)
	}
}

func (g *GrpcServer) Serve() error {
	address := g.Config.GetString("grpc.address") + ":" + g.Config.GetString("grpc.port")
	listener, err := net.Listen(g.Config.GetString("grpc.network"), address)
	if err != nil {
		return err
	}

	return g.Server.Serve(listener)
}

func (g *GrpcServer) proxyService(wg *sync.WaitGroup) {
	defer wg.Done()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	customMarshaller := &runtime.JSONPb{
		OrigName:     true,
		EmitDefaults: true, // disable 'omitempty'
	}
	muxOpt := runtime.WithMarshalerOption(runtime.MIMEWildcard, customMarshaller)
	mux := runtime.NewServeMux(muxOpt)

	opts := []grpc.DialOption{grpc.WithInsecure()}
	grpcAddress := g.Config.GetString("grpc.address") + ":" + g.Config.GetString("grpc.port")
	err := grpcHandlerAddressV1.RegisterAddressHandlerHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		g.Logger.Fatal("error reg address endpoint", err)
	}
	err = grpcHandlerHealth.RegisterHealthHandlerHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		g.Logger.Fatal("error reg health endpoint", err)
	}
	err = grpcHandlerVersion.RegisterVersionHandlerHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		g.Logger.Fatal("error reg version endpoint", err)
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	gatewayAddress := g.Config.GetString("grpc.gateway.address") + ":" + g.Config.GetString("grpc.gateway.port")
	g.Logger.Info("Start Http server on: " + gatewayAddress)
	g.Logger.Fatal(http.ListenAndServe(gatewayAddress, mux).Error())
}

func serverInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)
	var xRequestId string
	for i, v := range md {
		if i == "x-request-id" {
			xRequestId = v[0]
		}
	}

	globalLogger.WithFields(interfaces.LoggerFields{
		"x-request-id": xRequestId,
		"method":       info.FullMethod,
		"request":      req,
	}).Info("Request")

	// Проверка валидации, jwt, если потребуется
	/*if info.FullMethod != "/proto.EventStoreService/GetJWT" {
		if err := authorize(ctx); err != nil {
			return nil, err
		}
	}*/

	// Calls the handler
	h, err := handler(ctx, req)

	globalLogger.WithFields(interfaces.LoggerFields{
		"x-request-id": xRequestId,
		"method":       info.FullMethod,
		"response":     h,
		"time":         time.Since(start),
	}).Info("Response")

	return h, err
}

package grpc

import (
	"context"
	"github.com/GarinAG/gofias/domain/address/service"
	versionService "github.com/GarinAG/gofias/domain/version/service"
	grpcHandlerFiasV1 "github.com/GarinAG/gofias/infrastructure/persistence/grpc/dto/v1/fias"
	handlers "github.com/GarinAG/gofias/infrastructure/persistence/grpc/handler"
	"github.com/GarinAG/gofias/infrastructure/persistence/swagger/handler"
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

var Version = "3.0.1"

// Объект GRPC-сервера
type GrpcServer struct {
	Server         *grpc.Server                   // GRPC-сервер
	Logger         interfaces.LoggerInterface     // Логгер
	Config         interfaces.ConfigInterface     // Конфигурации
	AddressService *service.AddressService        // Сервис адресов
	HouseService   *service.HouseImportService    // Сервис домов
	VersionService *versionService.VersionService // Сервис версий
}

// Глобальный логгер для передачи в обработчик запросов
var globalLogger interfaces.LoggerInterface

// Глобальный логгер для передачи в обработчик запросов
var globalConfig interfaces.ConfigInterface

// Инициализация сервера
func NewGrpcServer(ctn *registry.Container) *GrpcServer {
	logger := ctn.Resolve("logger").(interfaces.LoggerInterface)
	config := ctn.Resolve("config").(interfaces.ConfigInterface)
	globalLogger = logger
	globalConfig = config

	defer func() {
		if r := recover(); r != nil {
			logger.WithFields(interfaces.LoggerFields{"error": r}).Panic("Program fatal error")
			os.Exit(1)
		}
	}()
	// Инициализация GRPC-сервера
	server := grpc.NewServer(grpc.UnaryInterceptor(serverInterceptor))
	// Регистрация обработчика адресов
	grpcHandlerFiasV1.RegisterAddressServiceServer(server,
		handlers.NewAddressHandler(
			ctn.Resolve("addressService").(*service.AddressService),
			ctn.Resolve("houseService").(*service.HouseService),
		))
	// Инициализация обработчика состояния приложения
	grpcHandlerFiasV1.RegisterHealthServiceServer(server, handlers.NewHealthHandler())
	// Инициализация обработчика версий
	grpcHandlerFiasV1.RegisterVersionServiceServer(server, handlers.NewVersionHandler(ctn.Resolve("versionService").(*versionService.VersionService), Version))
	reflection.Register(server)

	return &GrpcServer{
		Server:         server,
		Logger:         logger,
		Config:         config,
		AddressService: ctn.Resolve("addressService").(*service.AddressService),
		HouseService:   ctn.Resolve("houseImportService").(*service.HouseImportService),
		VersionService: ctn.Resolve("versionService").(*versionService.VersionService),
	}
}

// Запуск сервера
func (g *GrpcServer) Run() error {
	wg := sync.WaitGroup{}
	wg.Add(1)
	// Запускает GRPC-сервер
	go g.grpcService(&wg)
	if g.Config.GetConfig().Grpc.Gateway.Enable {
		wg.Add(1)
		// Запускает http-сервер
		go g.proxyService(&wg)
	}
	wg.Wait()
	return nil
}

// Запуск GRPC-сервера с обработчиком сигналов
func (g *GrpcServer) grpcService(wg *sync.WaitGroup) {
	defer wg.Done()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Слушает сигналы системы об остановке
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

// Запуск GRPC-сервера
func (g *GrpcServer) Serve() error {
	address := g.Config.GetConfig().Grpc.Address + ":" + g.Config.GetConfig().Grpc.Port
	listener, err := net.Listen(g.Config.GetConfig().Grpc.Network, address)
	if err != nil {
		return err
	}

	return g.Server.Serve(listener)
}

// Запуск http-сервера
func (g *GrpcServer) proxyService(wg *sync.WaitGroup) {
	defer wg.Done()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Настраивает вывод json
	customMarshaller := &runtime.JSONPb{
		OrigName:     true, // Возвращать оригинальные названия полей
		EmitDefaults: true, // Возвращать пустые значения
	}
	muxOpt := runtime.WithMarshalerOption(runtime.MIMEWildcard, customMarshaller)
	mux := runtime.NewServeMux(muxOpt)

	// Регистрирует обработчики запросов
	opts := []grpc.DialOption{grpc.WithInsecure()}
	grpcAddress := g.Config.GetConfig().Grpc.Address + ":" + g.Config.GetConfig().Grpc.Port
	// Регистрирует обработчик адресов
	err := grpcHandlerFiasV1.RegisterAddressServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		g.Logger.Fatal("error reg address endpoint", err)
	}
	// Регистрирует обработчик состояния приложения
	err = grpcHandlerFiasV1.RegisterHealthServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		g.Logger.Fatal("error reg health endpoint", err)
	}
	// Регистрирует обработчик версий
	err = grpcHandlerFiasV1.RegisterVersionServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		g.Logger.Fatal("error reg version endpoint", err)
	}
	// Регистрируем swagger
	err = handler.RegisterSwaggerHandlers(ctx, mux)
	if err != nil {
		g.Logger.Fatal("error reg swagger endpoint", err)
	}

	// Запускает http-сервер
	gatewayAddress := g.Config.GetConfig().Grpc.Gateway.Address + ":" + g.Config.GetConfig().Grpc.Gateway.Port
	g.Logger.Info("Start Http server on: " + gatewayAddress)

	g.Logger.Fatal(http.ListenAndServe(gatewayAddress, mux).Error())
}

// Инициализирует посредника запросов
func serverInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Выполняет действия перед выполнением запроса
	start := time.Now()
	// Читает метаданные
	md, _ := metadata.FromIncomingContext(ctx)
	var xRequestId string
	// Ищет ID запроса в заголовках
	for i, v := range md {
		if i == "x-request-id" {
			xRequestId = v[0]
		}
	}

	if globalConfig.GetConfig().Grpc.SaveRequest {
		globalLogger.WithFields(interfaces.LoggerFields{
			"x-request-id": xRequestId,
			"method":       info.FullMethod,
			"request":      req,
		}).Info("Request")
	}

	// Проверка валидации, jwt, если потребуется
	/*if info.FullMethod != "/proto.EventStoreService/GetJWT" {
		if err := authorize(ctx); err != nil {
			return nil, err
		}
	}*/

	// Исполняет запрос
	h, err := handler(ctx, req)

	// Выполняет действия после выполнения запроса
	if globalConfig.GetConfig().Grpc.SaveResponse {
		globalLogger.WithFields(interfaces.LoggerFields{
			"x-request-id": xRequestId,
			"method":       info.FullMethod,
			"response":     h,
			"time":         time.Since(start),
		}).Info("Response")
	}

	return h, err
}

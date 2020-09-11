package main

import (
	"flag"
	"fmt"
	grpc2 "github.com/GarinAG/gofias/infrastructure/persistence/grpc"
	"github.com/GarinAG/gofias/infrastructure/registry"
	"github.com/GarinAG/gofias/interfaces"
	"runtime"
)

// Основная функция запуска grpc сервера
func main() {
	// Чтение переданных в консоль флагов и установка максимального количества процессов
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Инициализация контейнера зависимостей
	ctn, err := registry.NewContainer("grpc")
	if err != nil {
		panic(fmt.Sprintf("Failed to init container: %v", err))
	}

	// Запуск grpc сервера
	app := grpc2.NewGrpcServer(ctn)
	if err := app.Run(); err != nil {
		app.Logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Program fatal error")
	}
}

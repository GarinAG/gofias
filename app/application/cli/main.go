package main

import (
	"flag"
	"fmt"
	addressCli "github.com/GarinAG/gofias/domain/address/delivery/cli"
	osmCli "github.com/GarinAG/gofias/domain/osm/delivery/cli"
	versionCli "github.com/GarinAG/gofias/domain/version/delivery/cli"
	indexCli "github.com/GarinAG/gofias/infrastructure/persistence/address/elastic/delivery/cli"
	cli2 "github.com/GarinAG/gofias/infrastructure/persistence/cli"
	"github.com/GarinAG/gofias/infrastructure/registry"
	"github.com/GarinAG/gofias/interfaces"
	"github.com/GarinAG/gofias/util"
	"runtime"
)

// Основная функция запуска консольного приложения по обновлению данных
func main() {
	// Чтение переданных в консоль флагов и установка максимального количества процессов
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Инициализация контейнера зависимостей
	ctn, err := registry.NewContainer("cli")
	if err != nil {
		panic(fmt.Sprintf("Failed to init container: %v", err))
	}

	// Инициализация глобальной переменно вывода информации в консоль
	util.CanPrintProcess = ctn.Resolve("config").(interfaces.ConfigInterface).GetBool("process.print")

	// Инициализация приложения
	app := cli2.NewApp(ctn)
	addressCli.RegisterImportCliEndpoint(app)
	indexCli.RegisterIndexCliEndpoint(app)
	versionCli.RegisterVersionCliEndpoint(app)
	osmCli.RegisterOsmCliEndpoint(app)

	// Запуск приложения
	if err := app.Run(); err != nil {
		app.Logger.WithFields(interfaces.LoggerFields{"error": err}).Fatal("Program fatal error")
	}
}

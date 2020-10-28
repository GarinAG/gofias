package cli

import (
	cli2 "github.com/GarinAG/gofias/infrastructure/persistence/cli"
	"github.com/urfave/cli/v2"
)

// Регистрация команды импорта
func RegisterImportCliEndpoint(app *cli2.App) {
	h := NewHandler(app.ImportService, app.OsmService, app.Logger)
	app.Server.Commands = append(app.Server.Commands, &cli.Command{
		Name:  "update",
		Usage: "Run fias import",
		Flags: []cli.Flag{
			// Флаг пропуска домов
			&cli.BoolFlag{
				Name:  "skip-houses",
				Value: false,
				Usage: "Skip houses import",
			},
			// Флаг запрета очистки временной директории
			&cli.BoolFlag{
				Name:  "skip-clear",
				Value: false,
				Usage: "Skip clear tmp folder on startup",
			},
			// Флаг запрета загрузки OSM данных после индексации
			&cli.BoolFlag{
				Name:  "skip-osm",
				Value: false,
				Usage: "Skip osm update",
			},
		},
		Action: func(c *cli.Context) error {
			app.ImportService.SkipHouses = c.Bool("skip-houses")
			app.ImportService.SkipClear = c.Bool("skip-clear")
			app.ImportService.SkipOsm = c.Bool("skip-osm")

			h.CheckUpdates(app.FiasApiService, app.VersionService)
			return nil
		},
	})
}

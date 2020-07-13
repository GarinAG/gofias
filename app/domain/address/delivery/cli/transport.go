package cli

import (
	server "github.com/GarinAG/gofias/server/cli"
	"github.com/urfave/cli/v2"
)

func RegisterImportCliEndpoint(app *server.App) {
	h := NewHandler(app.ImportService, app.Logger)
	app.Server.Commands = append(app.Server.Commands, &cli.Command{
		Name:  "update",
		Usage: "Run fias import",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-houses",
				Value: false,
				Usage: "Skip houses import",
			},
			&cli.BoolFlag{
				Name:  "skip-clear",
				Value: false,
				Usage: "Skip clear tmp folder on startup",
			},
		},
		Action: func(c *cli.Context) error {
			app.ImportService.SkipHouses = c.Bool("skip-houses")
			app.ImportService.SkipClear = c.Bool("skip-clear")

			h.CheckUpdates(app.FiasApiService, app.VersionService)
			return nil
		},
	})
}

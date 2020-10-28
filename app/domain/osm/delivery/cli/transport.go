package cli

import (
	cli2 "github.com/GarinAG/gofias/infrastructure/persistence/cli"
	"github.com/urfave/cli/v2"
)

// Регистрация команды импорта местоположений
func RegisterOsmCliEndpoint(app *cli2.App) {
	h := NewHandler(app.OsmService)
	app.Server.Commands = append(app.Server.Commands, &cli.Command{
		Name:  "osm-update",
		Usage: "UpdateFromExistItem geo-data",
		Action: func(c *cli.Context) error {
			h.Update()
			return nil
		},
	})
}

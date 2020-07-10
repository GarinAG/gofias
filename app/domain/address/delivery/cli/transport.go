package cli

import (
	server "github.com/GarinAG/gofias/server/cli"
	"github.com/urfave/cli"
)

func RegisterCliEndpoints(app *server.App) {
	h := NewHandler(app.ImportService, app.Logger)
	app.Server.Commands = []cli.Command{
		{
			Name:  "checkupdates",
			Usage: "fias run full import or delta's",
			Action: func(c *cli.Context) {
				h.CheckUpdates(app.FiasApiService, app.VersionService)
			},
		},
	}
}

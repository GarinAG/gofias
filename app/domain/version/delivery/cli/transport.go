package cli

import (
	server "github.com/GarinAG/gofias/server/cli"
	"github.com/urfave/cli"
)

func RegisterCliEndpoints(app *server.App) {
	h := NewHandler(*app.VersionService)
	app.Server.Commands = []cli.Command{
		{
			Name:  "version",
			Usage: "fias version",
			Action: func(c *cli.Context) {
				h.GetVersionInfo()
			},
		},
	}
}

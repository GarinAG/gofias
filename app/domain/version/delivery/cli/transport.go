package cli

import (
	server "github.com/GarinAG/gofias/server/cli"
	"github.com/urfave/cli/v2"
)

func RegisterCliEndpoints(app *server.App) {
	h := NewHandler(*app.VersionService)
	app.Server.Commands = append(app.Server.Commands, &cli.Command{
		Name:  "version",
		Usage: "Get current fias version",
		Action: func(c *cli.Context) error {
			h.GetVersionInfo()
			return nil
		},
	})
}

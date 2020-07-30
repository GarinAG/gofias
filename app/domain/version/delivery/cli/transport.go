package cli

import (
	cli2 "github.com/GarinAG/gofias/infrastructure/persistence/cli"
	"github.com/urfave/cli/v2"
)

func RegisterVersionCliEndpoint(app *cli2.App) {
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

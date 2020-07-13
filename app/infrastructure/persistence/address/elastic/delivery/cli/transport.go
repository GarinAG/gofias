package cli

import (
	server "github.com/GarinAG/gofias/server/cli"
	"github.com/urfave/cli/v2"
)

func RegisterIndexCliEndpoint(app *server.App) {
	h := NewHandler(app.ImportService, app.Logger)
	app.Server.Commands = append(app.Server.Commands, &cli.Command{
		Name:  "index",
		Usage: "Run fias elastic index",
		Action: func(c *cli.Context) error {
			h.importService.IsFull = true
			h.Index()
			return nil
		},
	})
}

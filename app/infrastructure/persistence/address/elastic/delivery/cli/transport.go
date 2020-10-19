package cli

import (
	cli2 "github.com/GarinAG/gofias/infrastructure/persistence/cli"
	"github.com/urfave/cli/v2"
)

// Регистрация основной команды индексации
func RegisterIndexCliEndpoint(app *cli2.App) {
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

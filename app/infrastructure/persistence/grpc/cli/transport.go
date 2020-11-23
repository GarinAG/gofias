package cli

import (
	cli2 "github.com/GarinAG/gofias/infrastructure/persistence/cli"
	"github.com/urfave/cli/v2"
)

// Регистрация команды импорта местоположений
func RegisterGrpcCliEndpoint(app *cli2.App) {
	h := NewHandler(app.Container)
	app.Server.Commands = append(app.Server.Commands, &cli.Command{
		Name:  "grpc",
		Usage: "Run grpc-server",
		Action: func(c *cli.Context) error {
			h.Run()
			return nil
		},
	})
}

package cli

import (
	addressCli "github.com/GarinAG/gofias/domain/address/delivery/cli"
	versionCli "github.com/GarinAG/gofias/domain/version/delivery/cli"
	"github.com/GarinAG/gofias/server/cli"
)

func main() {
	app := cli.NewApp()
	addressCli.RegisterCliEndpoints(app)
	versionCli.RegisterCliEndpoints(app)

	if err := app.Run(); err != nil {
		app.Logger.Fatal(err.Error())
	}
}

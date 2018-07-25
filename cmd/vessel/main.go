package main

import (
	"log"
	"os"

	"github.com/josebalius/vessel"
	"github.com/urfave/cli"
)

// pipelines
// mobile.yaml
// business.yaml

// functions
// mobile/mapper
// mobile/stopFinder

// config.yaml

// vessel deploy # deployes pipelines and functions

func main() {
	app := cli.NewApp()
	app.Name = "Vessel"
	app.Usage = "AWS Data Pipelines"

	app.Commands = cli.Commands{
		{
			Name:    "deploy",
			Aliases: []string{"d"},
			Usage:   "Deploys all pipelines and functions",
			Action:  deploy,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func deploy(c *cli.Context) error {
	if err := vessel.Deploy(); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

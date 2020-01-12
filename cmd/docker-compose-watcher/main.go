package main

import (
	"docker-compose-watcher/internal/business"
	"os"

	"github.com/urfave/cli/v2"
)

const fileFlagName = "file"

func main() {
	app := &cli.App{
		Name:    "docker-compose-watcher",
		Usage:   "Tool for automatic rebuilds and restarts of containers in Docker Compose.",
		Version: "1.0.0",
		Authors: []*cli.Author{
			{
				Name:  "norgor",
				Email: "norgor@gmail.com",
			},
		},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    fileFlagName,
				Value:   cli.NewStringSlice("docker-compose.yml"),
				Aliases: []string{"f"},
				Usage:   "Path to the Docker Compose file",
			},
		},
		Action: func(ctx *cli.Context) error {
			c, err := business.NewComposeController(ctx.StringSlice(fileFlagName)...)
			defer c.Close()
			if err != nil {
				panic(err)
			}
			if err := c.Run(); err != nil {
				panic(err)
			}
			return nil
		},
	}
	app.Run(os.Args)
}

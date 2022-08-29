package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/shemanaev/inpxer/internal/config"
	"github.com/shemanaev/inpxer/internal/indexer"
	"github.com/shemanaev/inpxer/internal/server"
)

var (
	version = "dev"
)

type key int

const (
	contextConfig key = iota
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading config")
	}

	app := &cli.App{
		Version: version,
		Commands: []*cli.Command{
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "start server",
				Action:  serveAction,
			},
			{
				Name:    "import",
				Aliases: []string{"i"},
				Usage:   "import .inpx file",
				Action:  importAction,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "keep-deleted",
						Usage: "Keep records marked as \"Deleted\" in inp",
					},
				},
			},
		},
		Before: func(ctx *cli.Context) error {
			ctx.Context = context.WithValue(ctx.Context, contextConfig, cfg)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func importAction(ctx *cli.Context) error {
	cfg := ctx.Context.Value(contextConfig).(*config.MyConfig)
	fmt.Println("Starting import from:", ctx.Args().First())
	return indexer.Run(cfg, ctx.Args().First(), ctx.Bool("keep-deleted"))
}

func serveAction(ctx *cli.Context) error {
	cfg := ctx.Context.Value(contextConfig).(*config.MyConfig)
	fmt.Printf("Starting web server on: http://%s\n", cfg.Listen)
	return server.Run(cfg)
}

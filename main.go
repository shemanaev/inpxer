package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/shemanaev/inpxer/internal/config"
	"github.com/shemanaev/inpxer/internal/indexer"
	"github.com/shemanaev/inpxer/internal/server"
)

var (
	version = "dev"
	date    = "unknown"
)

type key int

const (
	contextConfig key = iota
)

func main() {
	app := &cli.App{
		Version: version,
		Commands: []*cli.Command{
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "start server",
				Action:  serveAction,
				Before:  loadConfig,
			},
			{
				Name:    "import",
				Aliases: []string{"i"},
				Usage:   "import .inpx file",
				Action:  importAction,
				Before:  loadConfig,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "keep-deleted",
						Usage: "Keep records marked as \"Deleted\" in inp",
					},
					&cli.BoolFlag{
						Name:  "partial",
						Usage: "Only add new records, never delete",
					},
				},
			},
		},
	}

	if buildDate, err := time.Parse(time.RFC3339, date); err == nil {
		server.BuildDate = buildDate
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func importAction(ctx *cli.Context) error {
	cfg := ctx.Context.Value(contextConfig).(*config.MyConfig)
	fmt.Println("Starting import from:", ctx.Args().First())
	return indexer.Run(cfg, ctx.Args().First(), ctx.Bool("keep-deleted"), ctx.Bool("partial"))
}

func serveAction(ctx *cli.Context) error {
	cfg := ctx.Context.Value(contextConfig).(*config.MyConfig)
	fmt.Printf("Starting web server on: http://%s\n", cfg.Listen)
	return server.Run(cfg)
}

func loadConfig(ctx *cli.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("error loading config %v", err)
	}

	ctx.Context = context.WithValue(ctx.Context, contextConfig, cfg)
	return nil
}

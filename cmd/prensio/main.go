package main

import (
	"context"
	"fmt"
	"os"

	"github.com/siddontang/go-log/log"
	"github.com/urfave/cli/v2"
	"github.com/vnvo/prensio/config"
	"github.com/vnvo/prensio/pipeline"
)

func main() {

	// prensio run --config=/path/to/conf.toml
	// prensio run --config-dir=/path/to/dir
	app := &cli.App{
		Name:        "prensio",
		HelpName:    "prensio",
		Usage:       "a CDC tool for mysql and kafka with flexible transfomations",
		Description: "a change data capture tool for mysql and kafka. It can apply a user supplied transformer logic on each event.",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "start the cdc pipeline(s) using provided configurations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "conf",
						Value:   "/etc/prensio/default-pipeline.toml",
						Aliases: []string{"c"},
						EnvVars: []string{"PRENSIO_PIPELINE_CONF"},
						Usage:   "configuration `FILE` for the pipline",
					},
				},
				Action: func(cCtx *cli.Context) error {
					fmt.Println("conf:", cCtx.String("conf}"))
					err := run(cCtx.String("conf"))
					if err != nil {
						msg := fmt.Sprintf("failed to run prensio: %s", err)
						return cli.Exit(msg, 1)
					}

					return nil
				},
			},
		},
	}

	app.Run(os.Args)
}

func run(cfgPath string) error {
	initLogger()

	conf, err := config.NewCDCConfig(cfgPath)
	if err != nil {
		return err
	}

	myPipeline := pipeline.NewCDCPipeline("first-pipeline", &conf)
	log.Infof("[%s] created.", "first-pipeline")

	err = myPipeline.Init()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	myPipeline.Run(ctx)

	return nil
}

func initLogger() {
	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevel = "debug"
	}

	log.SetLevelByName(logLevel)
}

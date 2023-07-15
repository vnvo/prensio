package main

import (
	"fmt"
	"os"

	"github.com/siddontang/go-log/log"
	"github.com/urfave/cli/v2"
	"github.com/vnvo/prensio/pipeline"
)

func main() {

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
					run(cCtx.String("conf"))
					return cli.Exit("", 0)
				},
			},
		},
	}

	app.Run(os.Args)
}

func run(cfgPath string) {
	initLogger()

	pm, err := pipeline.NewPipelineManager(cfgPath)
	if err != nil {
		panic(err)
	}

	pm.Start()

	/*

		//filepath.WalkDir(root string, fn WalkFunc)
		conf, err := config.NewCDCConfig(cfgPath)
		if err != nil {
			panic(err)
		}

			myPipeline := pipeline.NewCDCPipeline("first-pipeline", &conf)
			log.Infof("[%s] created.", "first-pipeline")

			err = myPipeline.Init()
			if err != nil {
				panic(err)
			}

			ctx := context.Background()
			myPipeline.Run(ctx)
	*/
}

func initLogger() {
	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevel = "debug"
	}

	log.SetLevelByName(logLevel)
}

package main

import (
	_ "bgfreshd/internal/filters"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"math/rand"
	"os"
	"time"
)

var Version = "0.1.0"
var CompiledAt = time.Now()

func RunCli() {
	var config string
	var verbose bool = false
	app := &cli.App{
		Name:     "bgfreshd",
		Version:  Version,
		Compiled: CompiledAt,
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Dylan Ross",
				Email: "pdylanross@gmail.com",
			},
		},
		Copyright: "(c) 2020 Dylan Ross",
		Flags: []cli.Flag{
			&cli.StringFlag{
				EnvVars: []string{
					"BGFRESHD_CONFIG",
				},
				Name:        "config, c",
				Usage:       "[REQUIRED] Load configuration from `FILE`",
				Required:    true,
				Destination: &config,
			},
			&cli.BoolFlag{
				EnvVars: []string{
					"BGFRESHD_VERBOSE",
				},
				Name:        "verbose",
				Destination: &verbose,
			},
		},
		Action: func(c *cli.Context) error {
			config := &BgFreshConfig{
				Config: config,
				Log:    logInit(verbose),
			}

			rand.Seed(time.Now().UTC().UnixNano())
			if svc, err := NewService(config); err != nil {
				return err
			} else {
				defer svc.Stop()
				if err = svc.Run(); err != nil {
					return err
				}
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		os.Exit(-1)
	}
}

func main() {
	RunCli()
}

func logInit(verbose bool) *logrus.Logger {
	log := logrus.New()

	log.SetFormatter(&logrus.TextFormatter{})

	if verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	return log
}

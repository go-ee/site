package main

import (
	"github.com/go-ee/emailbridge"
	"github.com/go-ee/utils/lg"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
)
import (
	"fmt"
	"github.com/urfave/cli/v2"
)

func main() {
	var server, root, configFile, targetFile, cors string
	var debug bool
	var port int

	app := cli.NewApp()
	app.Name = "site"
	app.Usage = "Web server for static web sites, like Hugo, with Email support"
	app.Version = "1.0"

	lg.LogrusTimeAsTimestampFormatter()

	app.Before = func(c *cli.Context) (err error) {
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}
		logrus.Debugf("execute %v", c.Command.Name)
		return
	}

	app.Commands = []*cli.Command{
		{
			Name:  "serve",
			Usage: "Serve static web site",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "debug",
					Aliases:     []string{"d"},
					Destination: &debug,
					Usage:       "Enable debug log level",
				}, &cli.StringFlag{
					Name:        "server",
					Aliases:     []string{"a"},
					Usage:       "Host for the HTTP server",
					Value:       "",
					Destination: &server,
				}, &cli.IntFlag{
					Name:        "port",
					Aliases:     []string{"p"},
					Usage:       "port for the HTTP server",
					Value:       8080,
					Destination: &port,
				}, &cli.StringFlag{
					Name:        "root",
					Aliases:     []string{"r"},
					Usage:       "Root directory for the HTTP server",
					Value:       ".",
					Destination: &root,
				}, &cli.StringFlag{
					Name:        "cors",
					Usage:       "CORS, Access-Control-Allow-Origin pattern",
					Destination: &cors,
				},
			},
			Action: func(c *cli.Context) (err error) {

				http.Handle("/", http.FileServer(http.Dir(root)))

				serverAddr := fmt.Sprintf("%v:%v", server, port)

				logrus.Infof("serve '%v' on 'http://%v'", root, serverAddr)
				err = http.ListenAndServe(serverAddr, nil)

				return
			},
		},
		{
			Name:  "serveWithEmailSupport",
			Usage: "with email support",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "debug",
					Aliases:     []string{"d"},
					Destination: &debug,
					Usage:       "Enable debug log level",
				},
				&cli.StringFlag{
					Name:        "config",
					Aliases:     []string{"c"},
					Usage:       "EmailBridge config file",
					Value:       "config.xml",
					Destination: &configFile,
				},
			},
			Action: func(c *cli.Context) (err error) {
				var config emailbridge.Config
				if err = emailbridge.LoadConfig(configFile, &config); err == nil {
					if _, err = emailbridge.NewEmailBridge(&config, http.DefaultServeMux); err == nil {

						serverAddr := fmt.Sprintf("%v:%v", config.Server, config.Port)

						linkHost := config.Server
						if linkHost == "" {
							linkHost = "localhost"
						}

						logrus.Infof("serve '%v' on 'http://%v:%v' with email support '%v'",
							root, linkHost, port, config.Sender.Email)

						fileServer := http.FileServer(http.Dir(config.Root))

						http.Handle("/", fileServer)
						err = http.ListenAndServe(serverAddr, nil)
					}
				}
				return
			},
		},
		{
			Name:  "config",
			Usage: "Generate default config file",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "target",
					Aliases:     []string{"t"},
					Usage:       "Config target file name to generate",
					Value:       "config.yml",
					Destination: &targetFile,
				},
			},
			Action: func(c *cli.Context) (err error) {
				config := emailbridge.BuildDefault()
				config.Routes.Prefix = "_api/"
				err = emailbridge.WriteConfig(targetFile, config)
				return
			},
		},
		{
			Name:
			"markdown",
			Usage: "Generate markdown help file",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "target",
					Aliases:     []string{"t"},
					Usage:       "Markdown target file name to generate",
					Value:       "site.md",
					Destination: &targetFile,
				},
			},
			Action: func(c *cli.Context) (err error) {
				if markdown, err := app.ToMarkdown(); err == nil {
					err = ioutil.WriteFile(targetFile, []byte(markdown), 0)
				} else {
					logrus.Infof("%v", err)
				}
				return
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Warn("exit because of error.")
	}
}

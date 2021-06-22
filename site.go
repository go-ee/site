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

func main()  {
	app := NewCli()

	if err := app.Run(os.Args); err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Warn("exit because of error.")
	}
}

type Cli struct {
	*cli.App
	server, root, configFile, targetFile, cors string
	debug                                      bool
	port                                       int
}

func NewCli() (ret *Cli) {
	ret = &Cli{}
	ret.init()
	return
}

func (o *Cli) init() {
	o.App = cli.NewApp()
	o.Name = "site"
	o.Usage = "Web server for static web sites, like Hugo, with Email support"
	o.Version = "1.0"

	lg.LogrusTimeAsTimestampFormatter()

	o.Before = func(c *cli.Context) (err error) {
		if o.debug {
			logrus.SetLevel(logrus.DebugLevel)
		}
		logrus.Debugf("execute %v", c.Command.Name)
		return
	}

	o.Commands = []*cli.Command{
		{
			Name:  "serve",
			Usage: "Serve static web site",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "debug",
					Aliases:     []string{"d"},
					Destination: &o.debug,
					Usage:       "Enable debug log level",
				}, &cli.StringFlag{
					Name:        "server",
					Aliases:     []string{"a"},
					Usage:       "Host for the HTTP server",
					Value:       "",
					Destination: &o.server,
				}, &cli.IntFlag{
					Name:        "port",
					Aliases:     []string{"p"},
					Usage:       "port for the HTTP server",
					Value:       7070,
					Destination: &o.port,
				}, &cli.StringFlag{
					Name:        "root",
					Aliases:     []string{"r"},
					Usage:       "Root directory for the HTTP server",
					Value:       ".",
					Destination: &o.root,
				}, &cli.StringFlag{
					Name:        "cors",
					Usage:       "CORS, Access-Control-Allow-Origin pattern",
					Destination: &o.cors,
				},
			},
			Action: func(c *cli.Context) (err error) {

				http.Handle("/", http.FileServer(http.Dir(o.root)))

				serverAddr := fmt.Sprintf("%v:%v", o.server, o.port)
				linkHost := o.server
				if linkHost == "" || linkHost == "0.0.0.0" {
					linkHost = "127.0.0.1"
				}

				logrus.Infof("serve '%v' on 'http://%v:%v'", o.root, linkHost, o.port)
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
					Destination: &o.debug,
					Usage:       "Enable debug log level",
				},
				&cli.StringFlag{
					Name:        "config",
					Aliases:     []string{"c"},
					Usage:       "EmailBridge config file",
					Value:       "config.yml",
					Destination: &o.configFile,
				},
			},
			Action: func(c *cli.Context) (err error) {
				var config emailbridge.Config
				if err = emailbridge.ConfigLoad(o.configFile, &config); err == nil {
					if _, err = emailbridge.NewEmailBridge(&config, http.DefaultServeMux); err == nil {

						serverAddr := fmt.Sprintf("%v:%v", config.Server, config.Port)

						linkHost := config.Server
						if linkHost == "" || linkHost == "0.0.0.0" {
							linkHost = "127.0.0.1"
						}

						logrus.Infof("serve '%v' on 'http://%v:%v' with email support '%v'",
							o.root, linkHost, o.port, config.EngineConfig.Sender.Email)

						fileServer := http.FileServer(http.Dir(config.StaticFolder))

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
					Destination: &o.targetFile,
				},
			},
			Action: func(c *cli.Context) (err error) {
				config := emailbridge.BuildDefault()
				config.Routes.Prefix = "_api/"
				err = config.WriteConfig(o.targetFile)
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
					Destination: &o.targetFile,
				},
			},
			Action: func(c *cli.Context) (err error) {
				if markdown, err := o.ToMarkdown(); err == nil {
					err = ioutil.WriteFile(o.targetFile, []byte(markdown), 0)
				} else {
					logrus.Infof("%v", err)
				}
				return
			},
		},
	}
}

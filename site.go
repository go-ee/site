package main

import (
	"github.com/go-ee/utils/email"
	"github.com/go-ee/utils/lg"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)
import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	var emailAddress, smtpLogin, smtpPassword, smtpHost, address, root, targetFile string
	var port, smtpPort int
	var debug bool

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

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "verbose",
			Destination: &debug,
			Usage:       "Enable debug log level",
		}, &cli.StringFlag{
			Name:        "address",
			Aliases:     []string{"a"},
			Usage:       "Host for the HTTP server",
			Value:       "0.0.0.0",
			Destination: &address,
		}, &cli.IntFlag{
			Name:        "port",
			Aliases:     []string{"p"},
			Usage:       "port for the HTTP server",
			Value:       8080,
			Destination: &port,
		}, &cli.StringFlag{
			Name:        "directory",
			Aliases:     []string{"d"},
			Usage:       "Root directory for the HTTP server",
			Value:       ".",
			Destination: &root,
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:  "serve",
			Usage: "Serve static web site",
			Action: func(c *cli.Context) (err error) {
				log.Print(fmt.Sprintf("serve %v on %v:%v", root, address, port))
				http.Handle("/", http.FileServer(http.Dir(root)))
				err = http.ListenAndServe(address, nil)
				return
			},
			Subcommands: []*cli.Command{
				{
					Name:  "emailSupport",
					Usage: "with email support",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:        "email",
							Usage:       "Email address of sender",
							Required:    true,
							Destination: &emailAddress,
						}, &cli.StringFlag{
							Name:        "smtpLogin",
							Usage:       "SMTP login",
							Required:    true,
							Destination: &smtpLogin,
						}, &cli.StringFlag{
							Name:        "smtpPassword",
							Usage:       "SMTP password",
							Required:    true,
							Destination: &smtpPassword,
						}, &cli.StringFlag{
							Name:        "smtpHost",
							Usage:       "SMTP Server Host",
							Value:       "smtp.gmail.com",
							Destination: &smtpHost,
						}, &cli.IntFlag{
							Name:        "smtpPort",
							Usage:       "Sender Server port",
							Value:       587,
							Destination: &smtpPort,
						},
					},
					Action: func(c *cli.Context) (err error) {
						serverAddr := fmt.Sprintf("%v:%v", address, port)
						log.Print(fmt.Sprintf("serve '%v' on 'http://%v' with email support '%v'",
							root, serverAddr, emailAddress))
						http.Handle("/_api/email", EmailSupport(
							email.NewSender(emailAddress, smtpLogin, smtpPassword, smtpHost, smtpPort)))
						http.Handle("/", http.FileServer(http.Dir(root)))
						err = http.ListenAndServe(serverAddr, nil)
						return
					},
				},
			},
		},
		{
			Name:  "markdown",
			Usage: "Generate markdown help file",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "target",
					Aliases:     []string{"-t"},
					Usage:       "Markdown target file name to generate",
					Required:    true,
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

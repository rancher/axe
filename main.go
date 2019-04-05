//go:generate go run types/codegen/cleanup/main.go
//go:generate go run types/codegen/main.go

package main

import (
	"github.com/rancher/axe/throwing/k8s"
	"github.com/rancher/axe/throwing/rio"
	"os"

	"github.com/rancher/axe/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func init() {
	file, err := os.Create("throwing.logs")
	if err != nil {
		panic(err)
	}
	logrus.SetOutput(file)
}

func main() {
	app := cli.NewApp()
	app.Name = "throwing"
	app.Version = version.VERSION
	app.Usage = "throwing needs help!"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "kubeconfig",
			EnvVar: "KUBECONFIG",
			Value:  "${HOME}/.kube/config",
		},
		cli.StringFlag{
			Name:   "blade",
			Value:  "rio",
		},
	}
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if c.String("blade") == "rio"{
		return rio.Start(c)
	} else if c.String("blade") == "k8s"{
		return k8s.Start(c)
	}

	logrus.Warnf("You have not register a blade called %s. Exiting...", c.String("blade"))
	return nil
}

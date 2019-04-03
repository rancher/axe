//go:generate go run types/codegen/cleanup/main.go
//go:generate go run types/codegen/main.go

package main

import (
	"os"

	"github.com/rancher/axe/axe"
	"github.com/rancher/axe/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	file, err := os.Create("axe.logs")
	if err != nil {
		panic(err)
	}
	logrus.SetOutput(file)
}

func main() {
	app := cli.NewApp()
	app.Name = "axe"
	app.Version = version.VERSION
	app.Usage = "axe needs help!"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "kubeconfig",
			EnvVar: "KUBECONFIG",
			Value:  "${HOME}/.kube/config",
		},
	}
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	kubeconfig := c.String("kubeconfig")
	os.Setenv("KUBECONFIG", kubeconfig)

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}
	clientset := kubernetes.NewForConfigOrDie(restConfig)

	app := axe.NewAppView(clientset)
	if err := app.Init(); err != nil {
		return err
	}
	return app.Run()
}

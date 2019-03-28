package main

import (
	"github.com/rancher/axe/types/apis/some.api.group/v1"
	"github.com/rancher/norman/generator"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := generator.DefaultGenerate(v1.Schemas, "github.com/rancher/axe/types", false, nil); err != nil {
		logrus.Fatal(err)
	}
}

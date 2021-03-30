package util

import "github.com/sirupsen/logrus"

type HelmCmd struct {
	Cmd
}

func NewHelmCommand(l *logrus.Entry) *HelmCmd {
	return &HelmCmd{*NewCommand(l, "helm")}
}

func (c HelmCmd) UpdateDependency(chart, chartPath string) (string, error) {
	c.l.Infof("Running helm dependency update for chart: %v", chart)
	return c.Base().Args("dependency", "update", chartPath).Run()
}

func (c HelmCmd) Package(chart, chartPath, destPath string) (string, error) {
	c.l.Infof("Running helm package for chart: %v", chart)
	return c.Base().Args("package", chartPath, "--destination", destPath).Run()
}

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
	return c.Args("dependency", "update", chartPath, "--skip-refresh").Run()
}

func (c HelmCmd) Install(chart, chartPath string) (string, error) {
	c.l.Infof("Running helm install for chart: %v", chart)
	return c.Args("upgrade", chart, chartPath, "--install").Run()
}

func (c HelmCmd) Uninstall(chart string) (string, error) {
	c.l.Infof("Running helm uninstall for chart: %v", chart)
	return c.Args("uninstall", chart).Run()
}

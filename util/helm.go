package util

import "github.com/sirupsen/logrus"

func NewHelmCommand(l *logrus.Entry) (*CliCommand, error) {
	return NewCliCommand(l, "helm")
}

func (cc CliCommand) UpdateDependency(chart, chartPath string) (string, error) {
	cc.l.Infof("Running helm dependency update for chart: %v", chart)
	cmd := append(cc.GetCommand(), "dependency", "update", chartPath)
	return Command(cc.l, cmd...).Run()
}

func (cc CliCommand) Install(chart, chartPath string) (string, error) {
	cc.l.Infof("Running helm install for chart: %v", chart)
	cmd := append(cc.GetCommand(),
		"upgrade", chart, chartPath,
		"--install")
	return Command(cc.l, cmd...).Run()
}

func (cc CliCommand) Uninstall(chart string) (string, error) {
	cc.l.Infof("Running helm uninstall for chart: %v", chart)
	cmd := append(cc.GetCommand(),
		"uninstall", chart)
	return Command(cc.l, cmd...).Run()
}

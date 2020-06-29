package util

import "github.com/sirupsen/logrus"

type HelmCommand struct {
	*CliCommand
	l         *logrus.Entry
	namespace string
}

func NewHelmCommand(l *logrus.Entry, namespace string) *HelmCommand {
	return &HelmCommand{&CliCommand{"helm"}, l, namespace}
}

func (hc HelmCommand) UpdateDependency(chart, chartPath string) (string, error) {
	hc.l.Infof("Running helm dependency update for chart: %v", chart)
	cmd := []string{hc.name, "dependency", "update", chartPath}
	return Command(hc.l, cmd...).Run()
}

func (hc HelmCommand) Install(chart, chartPath string) (string, error) {
	hc.l.Infof("Running helm install for chart: %v", chart)
	cmd := []string{
		hc.name, "-n", hc.namespace,
		"upgrade", chart, chartPath,
		"--install",
	}
	return Command(hc.l, cmd...).Run()
}

func (hc HelmCommand) Uninstall(chart string) (string, error) {
	hc.l.Infof("Running helm uninstall for chart: %v", chart)
	cmd := []string{
		hc.name, "-n", hc.namespace,
		"uninstall", chart,
	}
	return Command(hc.l, cmd...).Run()
}

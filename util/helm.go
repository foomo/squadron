package util

import (
	"context"
)

type HelmCmd struct {
	Cmd
}

func NewHelmCommand() *HelmCmd {
	return &HelmCmd{*NewCommand("helm")}
}

func (c HelmCmd) UpdateDependency(ctx context.Context, chartPath string) (string, error) {
	return c.Base().Args("dependency", "update", chartPath).Run(ctx)
}

func (c HelmCmd) Package(ctx context.Context, chartPath, destPath string) (string, error) {
	return c.Base().Args("package", chartPath, "--destination", destPath).Run(ctx)
}

package util

type HelmCmd struct {
	Cmd
}

func NewHelmCommand() *HelmCmd {
	return &HelmCmd{*NewCommand("helm")}
}

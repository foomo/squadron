package actions

import (
	"github.com/foomo/squadron"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	pushCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
	pushCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "builds or rebuilds units")
	pushCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	pushCmd.Flags().StringArrayVar(&flagBuildArgs, "build-args", nil, "additional docker buildx build args")
	pushCmd.Flags().StringArrayVar(&flagPushArgs, "push-args", nil, "additional docker push args")
	pushCmd.Flags().StringSliceVar(&flagTags, "tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
}

var pushCmd = &cobra.Command{
	Use:     "push [SQUADRON] [UNIT...]",
	Short:   "pushes the squadron or given units",
	Example: "  squadron push storefinder frontend backend --namespace demo --build",
	RunE: func(cmd *cobra.Command, args []string) error {
		sq := squadron.New(cwd, flagNamespace, flagFiles)

		if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
			return errors.Wrap(err, "failed to merge config files")
		}

		squadronName, unitNames := parseSquadronAndUnitNames(args)
		if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, flagTags); err != nil {
			return errors.Wrap(err, "failed to filter config")
		}

		if err := sq.RenderConfig(cmd.Context()); err != nil {
			return errors.Wrap(err, "failed to render config")
		}

		if flagBuild {
			if err := sq.Build(cmd.Context(), flagBuildArgs, flagParallel); err != nil {
				return errors.Wrap(err, "failed to build units")
			}
		}

		return sq.Push(cmd.Context(), flagPushArgs, flagParallel)
	},
}

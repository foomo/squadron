package actions

import (
	"github.com/foomo/squadron"
	"github.com/spf13/cobra"
)

func init() {
	buildCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "pushes built squadron units to the registry")
	buildCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	buildCmd.Flags().StringSliceVar(&flagTags, "tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	buildCmd.Flags().StringSliceVar(&flagBuildArgs, "build-args", nil, "additional docker buildx build args")
	buildCmd.Flags().StringSliceVar(&flagPushArgs, "push-args", nil, "additional docker push args")
}

var buildCmd = &cobra.Command{
	Use:     "build [SQUADRON.UNIT...]",
	Short:   "build or rebuild squadron units",
	Example: "  squadron build storefinder frontend backend",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		sq := squadron.New(cwd, "", flagFiles)

		if err := sq.MergeConfigFiles(); err != nil {
			return err
		}

		squadronName, unitNames := parseSquadronAndUnitNames(args)
		if err := sq.FilterConfig(squadronName, unitNames, flagTags); err != nil {
			return err
		}

		if err := sq.Build(cmd.Context(), flagBuildArgs, flagParallel); err != nil {
			return err
		}

		if flagPush {
			if err := sq.Push(cmd.Context(), flagPushArgs, flagParallel); err != nil {
				return err
			}
		}

		return nil
	},
}

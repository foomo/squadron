package actions

import (
	"fmt"
	"log"
	"os"

	"github.com/foomo/configurd"
	"github.com/spf13/cobra"
)

var (
	cnf     configurd.Configurd
	rootCmd = &cobra.Command{
		Use:   "cobra",
		Short: "A generator for Cobra based Applications",
		Long: `Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	}

	FlagTag string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&FlagTag, "tag", "t", "latest", "Specifies the image tag")
}

func Execute() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	cnf, err = configurd.New(dir)
	if err != nil {
		log.Fatal(err)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

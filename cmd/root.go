package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "az-docs",
	Short: "A way to generate compliance documentation based on your Azure tenant",
}

func Init() {
	rootCmd.SetOutput(os.Stdout)
	rootCmd.PersistentFlags().StringP("base-mgmt-group", "m", "", "The base management group to generate docs from")
	err := rootCmd.MarkPersistentFlagRequired("base-mgmt-group")
	if err != nil {
		panic(err)
	}
	addGenerate(rootCmd)
	addShow(rootCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

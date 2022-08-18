package cmd

import (
	"github.com/nepomuceno/az-docs/assets"
	"github.com/nepomuceno/az-docs/resources"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the entities available",
	RunE:  showAll,
}

func showAll(cmd *cobra.Command, args []string) error {
	baseMgmtGroup, err := cmd.Flags().GetString("base-mgmt-group")
	if err != nil {
		return err
	}
	cred := Login()
	diag := resources.NewAzDiagram(cred, baseMgmtGroup)
	err = diag.InitAll()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	docTemplate, err := assets.GetShowTemplate()
	if err != nil {
		return err
	}
	err = docTemplate.Execute(cmd.OutOrStdout(), diag)
	return err
}

func addShow(rootCommand *cobra.Command) {
	rootCmd.AddCommand(showCmd)
}

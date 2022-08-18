package cmd

import (
	"os"

	"github.com/nepomuceno/az-docs/assets"
	"github.com/nepomuceno/az-docs/resources"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a template for the Azure Policy CLI",
	RunE:  generateAll,
}

func generateAll(cmd *cobra.Command, args []string) error {
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
	cmd.Println("Generating documentation")
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	docTemplate, err := assets.GetDocTemplate()
	if err != nil {
		return err
	}
	err = docTemplate.Execute(f, diag)
	cmd.Println("Documentation generated")
	return err
}

func addGenerate(rootCommand *cobra.Command) {
	generateCmd.Flags().StringP("output", "o", "docs.md", "The output file to write the docs to")
	rootCmd.AddCommand(generateCmd)
}

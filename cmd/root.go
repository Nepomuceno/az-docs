package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy"
	"github.com/nepomuceno/az-docs/assets"
	"github.com/spf13/cobra"
)

type DocTemplate struct {
	Entities    map[string]*armmanagementgroups.EntityInfo
	Assignments map[string]*armpolicy.Assignment
}

func execCli(cmd *cobra.Command, args []string) error {
	cmd.Println("Getting policy definitions")
	cred := Login()
	_, _, err := getBuiltInPolicyDefinitions(cred)
	if err != nil {
		return err
	}
	cmd.Println("Getting managment groups and subs")
	entities, err := getEntities(cred, "bz-stg")
	if err != nil {
		return err
	}
	assignments := make(map[string]*armpolicy.Assignment, 0)
	cmd.Println("Getting assignments")
	for _, e := range entities {
		if *e.Type == "Microsoft.Management/managementGroups" {
			cmd.Printf("%s\n", *e.ID)
			managementGroupAssignments, err := getAssignmentsForManagmentGroup(cred, *e.Name)
			if err != nil {
				cmd.PrintErr(err)
				continue
			}
			for _, a := range managementGroupAssignments {
				assignments[*a.ID] = a
			}
		} else if *e.Type == "/subscriptions" {
			cmd.Printf("%s\n", *e.ID)
			subAssignments, err := getAssignments(cred, *e.Name)
			if err != nil {
				continue
			}
			for _, a := range subAssignments {
				assignments[*a.ID] = a
			}
		}
	}

	if len(assignments) > 0 {
		cmd.Printf("%d assignments found\n", len(assignments))
	} else {
		cmd.Printf("No assignments found\n")
	}

	cmd.Println("Generating documentation")
	f, err := os.Create("docs.md")
	if err != nil {
		return err
	}
	docTemplate, err := template.New("doc-template").Funcs(template.FuncMap{
		"mdlink": func(s *string) string {
			result := strings.ToLower(*s)
			result = strings.ReplaceAll(result, " ", "-")
			return result
		},
	}).Parse(assets.DocTemplate)
	if err != nil {
		return err
	}
	docTemplate.Execute(f, &DocTemplate{
		Entities:    entities,
		Assignments: assignments,
	})
	cmd.Println("Documentation generated")
	return nil
}

func getAssignmentsForManagmentGroup(cred *azidentity.DefaultAzureCredential, managmentGroupName string) ([]*armpolicy.Assignment, error) {
	client, err := armpolicy.NewAssignmentsClient(managmentGroupName, cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListForManagementGroupPager(managmentGroupName, &armpolicy.AssignmentsClientListForManagementGroupOptions{
		Filter: to.Ptr("atExactScope()"),
	})
	result := make([]*armpolicy.Assignment, 0)
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		result = append(result, page.Value...)
	}
	return result, nil
}

func getAssignments(cred *azidentity.DefaultAzureCredential, subsciptionID string) ([]*armpolicy.Assignment, error) {
	client, err := armpolicy.NewAssignmentsClient(subsciptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(nil)
	result := make([]*armpolicy.Assignment, 0)
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		result = append(result, page.Value...)
	}
	return result, nil
}

func getEntities(cred *azidentity.DefaultAzureCredential, basemg string) (map[string]*armmanagementgroups.EntityInfo, error) {
	client, err := armmanagementgroups.NewEntitiesClient(cred, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(&armmanagementgroups.EntitiesClientListOptions{
		Filter: to.Ptr(fmt.Sprintf("name eq '%s'", basemg)),
		Search: to.Ptr(armmanagementgroups.EntitySearchTypeParentAndFirstLevelChildren),
	})
	result := make(map[string]*armmanagementgroups.EntityInfo, 0)
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, e := range page.Value {
			result[*e.ID] = e
		}
	}
	for _, e := range result {
		if *e.Type == "Microsoft.Management/managementGroups" && *e.Name != basemg {
			childResult, err := getEntities(cred, *e.Name)
			if err != nil {
				return nil, err
			}
			for k, v := range childResult {
				result[k] = v
			}
		}
	}

	return result, nil
}

func getBuiltInPolicyDefinitions(cred *azidentity.DefaultAzureCredential) ([]*armpolicy.Definition, []*armpolicy.SetDefinition, error) {
	definitionsClient, err := armpolicy.NewDefinitionsClient("", cred, nil)
	if err != nil {
		return nil, nil, err
	}
	builtInPoliciesPager := definitionsClient.NewListBuiltInPager(nil)
	biultInPolicies := make([]*armpolicy.Definition, 0)
	for builtInPoliciesPager.More() {
		builtInPolicies, err := builtInPoliciesPager.NextPage(context.Background())
		if err != nil {
			return biultInPolicies, nil, err
		}
		biultInPolicies = append(biultInPolicies, builtInPolicies.Value...)
	}
	setDefinitionsClient, err := armpolicy.NewSetDefinitionsClient("", cred, nil)
	if err != nil {
		return biultInPolicies, nil, err
	}
	setPoliciesBuiltInPager := setDefinitionsClient.NewListBuiltInPager(nil)
	setPoliciesBuiltInResult := make([]*armpolicy.SetDefinition, 0)
	for setPoliciesBuiltInPager.More() {
		setPolicies, err := setPoliciesBuiltInPager.NextPage(context.Background())
		if err != nil {
			return biultInPolicies, setPoliciesBuiltInResult, err
		}
		setPoliciesBuiltInResult = append(setPoliciesBuiltInResult, setPolicies.Value...)
	}
	return biultInPolicies, setPoliciesBuiltInResult, nil
}

var rootCmd = &cobra.Command{
	Use:   "az-docs",
	Short: "A way to generate compliance documentation based on your Azure tenant",
	RunE:  execCli,
}

func Init() {
	rootCmd.Flags().StringP("base-mgmt-group", "m", "", "The base management group to generate docs from")
	rootCmd.Flags().StringP("output", "o", "docs.md", "The output file to write the docs to")

	err := rootCmd.MarkFlagRequired("base-mgmt-group")
	if err != nil {
		panic(err)
	}

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

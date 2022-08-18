package resources

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy"
)

type AzDiagram struct {
	BaseManagementGroup  string
	Credentials          *azidentity.DefaultAzureCredential
	Entities             map[string]*armmanagementgroups.EntityInfo
	Assignments          map[string]*armpolicy.Assignment
	Definitions          map[string]*armpolicy.Definition
	DefinitionSets       map[string]*armpolicy.SetDefinition
	UsedDefinitions      map[string]*armpolicy.Definition
	UsedDefinitionSets   map[string]*UsedDefinitionSet
	DefinitionScopes     map[string][]string
	DefinitionSetsScopes map[string][]string
}

type UsedDefinitionSet struct {
	armpolicy.SetDefinition
	Definitions []*armpolicy.Definition
}

func NewAzDiagram(cred *azidentity.DefaultAzureCredential, basemg string) *AzDiagram {
	diag := &AzDiagram{
		Credentials:          cred,
		BaseManagementGroup:  basemg,
		Entities:             make(map[string]*armmanagementgroups.EntityInfo),
		Assignments:          make(map[string]*armpolicy.Assignment),
		Definitions:          make(map[string]*armpolicy.Definition),
		DefinitionSets:       make(map[string]*armpolicy.SetDefinition),
		UsedDefinitions:      make(map[string]*armpolicy.Definition),
		UsedDefinitionSets:   make(map[string]*UsedDefinitionSet),
		DefinitionScopes:     make(map[string][]string),
		DefinitionSetsScopes: make(map[string][]string),
	}
	return diag
}

func (diag *AzDiagram) InitAll() error {
	err := diag.GetEntities()
	if err != nil {
		return err
	}
	err = diag.GetAssignments()
	if err != nil {
		return err
	}
	err = diag.GetDefinitions()
	if err != nil {
		return err
	}
	err = diag.GetUtilization()
	if err != nil {
		return err
	}
	return nil
}

func (diag *AzDiagram) GetAssignments() error {
	for _, e := range diag.Entities {
		var assignments []*armpolicy.Assignment
		var err error
		if *e.Type == "Microsoft.Management/managementGroups" {
			assignments, err = getAssignmentsForManagmentGroup(diag.Credentials, *e.Name)
		} else {
			assignments, err = getAssignments(diag.Credentials, *e.Name)
		}
		if err != nil {
			return err
		}
		for _, a := range assignments {
			diag.Assignments[*a.ID] = a
		}
	}
	return nil
}

func (diag *AzDiagram) GetUtilization() error {
	for _, a := range diag.Assignments {
		if def, ok := diag.Definitions[*a.Properties.PolicyDefinitionID]; ok {
			diag.UsedDefinitions[*def.ID] = def
			diag.DefinitionScopes[*def.ID] = append(diag.DefinitionScopes[*def.ID], *a.Properties.Scope)
		}
		if set, ok := diag.DefinitionSets[*a.Properties.PolicyDefinitionID]; ok {
			definitions := make([]*armpolicy.Definition, 0)
			for _, defRef := range set.Properties.PolicyDefinitions {
				if def, ok := diag.Definitions[*defRef.PolicyDefinitionID]; ok {
					definitions = append(definitions, def)
				}
			}
			diag.UsedDefinitionSets[*set.ID] = &UsedDefinitionSet{
				SetDefinition: *set,
				Definitions:   definitions,
			}
			diag.DefinitionSetsScopes[*set.ID] = append(diag.DefinitionSetsScopes[*set.ID], *a.Properties.Scope)
		}
	}
	return nil
}

func (diag *AzDiagram) GetEntities() error {
	entities, err := getEntities(diag.Credentials, diag.BaseManagementGroup)
	if err != nil {
		return err
	}
	diag.Entities = entities
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
func (diag *AzDiagram) GetDefinitions() error {
	var definitions []*armpolicy.Definition
	var definitionsSets []*armpolicy.SetDefinition
	var err error
	definitions, definitionsSets, err = getBuiltInPolicyDefinitions(diag.Credentials)
	if err != nil {
		return err
	}
	for _, d := range definitions {
		diag.Definitions[*d.ID] = d
	}
	for _, d := range definitionsSets {
		diag.DefinitionSets[*d.ID] = d
	}
	for _, a := range diag.Entities {

		if *a.Type == "Microsoft.Management/managementGroups" {
			definitions, definitionsSets, err = getDefinitionsForManagmentGroup(diag.Credentials, *a.Name)
		} else {
			definitions, definitionsSets, err = getDefinitions(diag.Credentials, *a.Name)
		}
		if err != nil {
			return err
		}
		for _, d := range definitions {
			diag.Definitions[*d.ID] = d
		}
		for _, d := range definitionsSets {
			diag.DefinitionSets[*d.ID] = d
		}

	}
	return nil
}

func getDefinitionsForManagmentGroup(cred *azidentity.DefaultAzureCredential, managmentGroupID string) ([]*armpolicy.Definition, []*armpolicy.SetDefinition, error) {
	definitionsClient, err := armpolicy.NewDefinitionsClient("", cred, nil)
	if err != nil {
		return nil, nil, err
	}
	definitionsPager := definitionsClient.NewListByManagementGroupPager(managmentGroupID, &armpolicy.DefinitionsClientListByManagementGroupOptions{
		Filter: to.Ptr("atExactScope()"),
	})
	definitions := make([]*armpolicy.Definition, 0)
	for definitionsPager.More() {
		page, err := definitionsPager.NextPage(context.Background())
		if err != nil {
			return nil, nil, err
		}
		definitions = append(definitions, page.Value...)
	}
	setsClient, err := armpolicy.NewSetDefinitionsClient("", cred, nil)
	if err != nil {
		return definitions, nil, err
	}
	setsPager := setsClient.NewListByManagementGroupPager(managmentGroupID, &armpolicy.SetDefinitionsClientListByManagementGroupOptions{
		Filter: to.Ptr("atExactScope()"),
	})
	sets := make([]*armpolicy.SetDefinition, 0)
	for setsPager.More() {
		page, err := setsPager.NextPage(context.Background())
		if err != nil {
			return definitions, nil, err
		}
		sets = append(sets, page.Value...)
	}
	return definitions, sets, nil
}

func getDefinitions(cred *azidentity.DefaultAzureCredential, subscriptionID string) ([]*armpolicy.Definition, []*armpolicy.SetDefinition, error) {
	definitionsClient, err := armpolicy.NewDefinitionsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, nil, err
	}
	definitionsPager := definitionsClient.NewListPager(&armpolicy.DefinitionsClientListOptions{
		Filter: to.Ptr("atExactScope()"),
	})
	definitions := make([]*armpolicy.Definition, 0)
	for definitionsPager.More() {
		page, err := definitionsPager.NextPage(context.Background())
		if err != nil {
			return nil, nil, err
		}
		definitions = append(definitions, page.Value...)
	}
	setsClient, err := armpolicy.NewSetDefinitionsClient(subscriptionID, cred, nil)
	if err != nil {
		return definitions, nil, err
	}
	setsPager := setsClient.NewListPager(&armpolicy.SetDefinitionsClientListOptions{
		Filter: to.Ptr("atExactScope()"),
	})
	sets := make([]*armpolicy.SetDefinition, 0)
	for setsPager.More() {
		page, err := setsPager.NextPage(context.Background())
		if err != nil {
			return definitions, nil, err
		}
		sets = append(sets, page.Value...)
	}
	return definitions, sets, nil
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

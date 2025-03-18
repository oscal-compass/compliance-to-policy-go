/*
Copyright 2023 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"fmt"
	"time"

	"github.com/defenseunicorns/go-oscal/src/pkg/uuid"
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	sigyaml "sigs.k8s.io/yaml"

	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
	provider "github.com/oscal-compass/compliance-to-policy-go/v2/policy"

	"k8s.io/apimachinery/pkg/util/sets"

	typeplacementdecision "github.com/oscal-compass/compliance-to-policy-go/v2/pkg/types/placementdecision"
	typepolicy "github.com/oscal-compass/compliance-to-policy-go/v2/pkg/types/policy"
	typeutils "github.com/oscal-compass/compliance-to-policy-go/v2/pkg/types/utils"
)

type ResultToOscal struct {
	policy             provider.Policy
	policyResultsDir   string
	policies           []*typepolicy.Policy
	policySets         []*typepolicy.PolicySet
	namespace          string
	policySetName      string
	placementDecisions []*typeplacementdecision.PlacementDecision
}

type Reason struct {
	ClusterName     string                         `json:"clusterName,omitempty" yaml:"clusterName,omitempty"`
	ComplianceState typepolicy.ComplianceState     `json:"complianceState,omitempty" yaml:"complianceState,omitempty"`
	Messages        []typepolicy.ComplianceHistory `json:"messages,omitempty" yaml:"messages,omitempty"`
}

type GenerationType string

const (
	GenerationTypeRaw          GenerationType = "raw"
	GenerationTypePolicyReport GenerationType = "policy-report"
)

func NewResultToOscal(pl provider.Policy, policyResultsDir, namespace, policySetName string) *ResultToOscal {
	r := ResultToOscal{
		policy:             pl,
		policyResultsDir:   policyResultsDir,
		policies:           []*typepolicy.Policy{},
		policySets:         []*typepolicy.PolicySet{},
		namespace:          namespace,
		policySetName:      policySetName,
		placementDecisions: []*typeplacementdecision.PlacementDecision{},
	}
	return &r
}

func (r *ResultToOscal) GenerateResults() (provider.PVPResult, error) {

	var policyList typepolicy.PolicyList
	if err := r.loadData("policies.policy.open-cluster-management.io.yaml", &policyList); err != nil {
		return provider.PVPResult{}, err
	}
	for idx := range policyList.Items {
		r.policies = append(r.policies, &policyList.Items[idx])
	}

	var policySetList typepolicy.PolicySetList
	if err := r.loadData("policysets.policy.open-cluster-management.io.yaml", &policySetList); err != nil {
		return provider.PVPResult{}, err
	}
	for idx := range policySetList.Items {
		r.policySets = append(r.policySets, &policySetList.Items[idx])
	}

	var placementDecisionLost typeplacementdecision.PlacementDecisionList
	if err := r.loadData("placementdecisions.cluster.open-cluster-management.io.yaml", &placementDecisionLost); err != nil {
		return provider.PVPResult{}, err
	}
	for idx := range placementDecisionLost.Items {
		r.placementDecisions = append(r.placementDecisions, &placementDecisionLost.Items[idx])
	}

	policySets := typeutils.FilterByAnnotation(r.policySets, pkg.ANNOTATION_COMPONENT_TITLE, r.policySetName)
	clusterNameSets := sets.NewString()
	var policySet *typepolicy.PolicySet
	if len(policySets) > 0 {
		policySet = policySets[0]
	}
	if policySet != nil {
		placements := []string{}
		for _, placement := range policySet.Status.Placement {
			placements = append(placements, placement.Placement)
		}
		for _, placement := range placements {
			placementDecision := typeutils.FindByNamespaceLabel(r.placementDecisions, policySet.Namespace, "cluster.open-cluster-management.io/placement", placement)
			for _, decision := range placementDecision.Status.Decisions {
				clusterNameSets.Insert(decision.ClusterName)
			}
		}
	}

	inventories := []oscalTypes.InventoryItem{}
	clusternameIndex := map[string]bool{}
	for _, policy := range r.policies {
		if policy.Namespace == r.namespace {
			for _, s := range policy.Status.Status {
				_, exist := clusternameIndex[s.ClusterName]
				if !exist {
					clusternameIndex[s.ClusterName] = true
					item := oscalTypes.InventoryItem{
						UUID: uuid.NewUUID(),
						Props: &[]oscalTypes.Property{
							{
								Name:  "cluster-name",
								Value: s.ClusterName,
								Ns:    extensions.TrestleNameSpace,
							},
						},
					}
					inventories = append(inventories, item)
				}
			}
		}
	}

	var observations []provider.ObservationByCheck
	for _, rule := range r.policy {
		for _, check := range rule.Checks {
			policyId := check.ID
			var policy *typepolicy.Policy

			if policySet != nil {
				policy = typeutils.FindByNamespaceName(r.policies, policySet.Namespace, policyId)
			}

			var subjects []provider.Subject
			if policy != nil {
				reasons := r.GenerateReasonsFromRawPolicies(*policy)
				for _, reason := range reasons {
					clusterName := "N/A"
					inventoryUuid := ""
					for _, inventory := range inventories {
						prop, ok := extensions.GetTrestleProp("cluster-name", *inventory.Props)
						if ok && prop.Value == reason.ClusterName {
							clusterName = prop.Value
							inventoryUuid = inventory.UUID
							break
						}
					}
					if inventoryUuid != "" {
						var message string
						if messageByte, err := sigyaml.Marshal(reason.Messages); err == nil {
							message = string(messageByte)
						} else {
							message = err.Error()
						}
						subject := provider.Subject{
							Type:        "resource",
							Title:       "Cluster Name: " + clusterName,
							ResourceID:  inventoryUuid,
							Result:      mapToPolicyResult(reason.ComplianceState),
							Reason:      message,
							EvaluatedOn: time.Now(),
						}
						subjects = append(subjects, subject)
					}
				}
			}

			props := []provider.Property{
				{
					Name:  "assessment-rule-id",
					Value: rule.Rule.ID,
				},
			}

			observation := provider.ObservationByCheck{
				Title:       rule.Rule.ID,
				CheckID:     policyId,
				Description: fmt.Sprintf("Observation of policy %s", policyId),
				Methods:     []string{"TEST-AUTOMATED"},
				Props:       props,
				Subjects:    subjects,
				Collected:   time.Now(),
			}
			observations = append(observations, observation)
		}
	}

	result := provider.PVPResult{
		ObservationsByCheck: observations,
	}

	return result, nil
}

func (r *ResultToOscal) GenerateReasonsFromRawPolicies(policy typepolicy.Policy) []Reason {
	reasons := []Reason{}
	for _, status := range policy.Status.Status {
		clusterName := status.ClusterName
		policyPerCluster := typeutils.FindByNamespaceName(r.policies, clusterName, policy.Namespace+"."+policy.Name)
		if policyPerCluster == nil {
			continue
		}
		messages := []typepolicy.ComplianceHistory{}
		for _, detail := range policyPerCluster.Status.Details {
			if len(detail.History) > 0 {
				messages = append(messages, detail.History[0])
			}
		}
		reasons = append(reasons, Reason{
			ClusterName:     clusterName,
			ComplianceState: status.ComplianceState,
			Messages:        messages,
		})
	}
	return reasons

}

func (r *ResultToOscal) loadData(path string, out interface{}) error {
	if err := pkg.LoadYamlFileToK8sTypedObject(r.policyResultsDir+"/"+path, &out); err != nil {
		return err
	}
	return nil
}

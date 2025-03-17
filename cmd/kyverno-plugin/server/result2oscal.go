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

	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	typepolr "sigs.k8s.io/wg-policy-prototypes/policy-report/pkg/api/wgpolicyk8s.io/v1beta1"

	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

type ResultToOscal struct {
	policy                  policy.Policy
	policyResultsDir        string
	policyReportList        *typepolr.PolicyReportList
	clusterPolicyReportList *typepolr.ClusterPolicyReportList
	policyList              *kyvernov1.PolicyList
	clusterPolicyList       *kyvernov1.ClusterPolicyList
}

type PolicyReportContainer struct {
	PolicyReports        []*typepolr.PolicyReport
	ClusterPolicyReports []*typepolr.ClusterPolicyReport
}

type PolicyResourceIndexContainer struct {
	PolicyResourceIndex PolicyResourceIndex
	ControlIds          []string
}

func NewResultToOscal(pl policy.Policy, policyResultsDir string) *ResultToOscal {
	r := ResultToOscal{
		policy:                  pl,
		policyResultsDir:        policyResultsDir,
		policyReportList:        &typepolr.PolicyReportList{},
		clusterPolicyReportList: &typepolr.ClusterPolicyReportList{},
		policyList:              &kyvernov1.PolicyList{},
		clusterPolicyList:       &kyvernov1.ClusterPolicyList{},
	}
	return &r
}

func (r *ResultToOscal) retrievePolicyReportResults(name string) []*typepolr.PolicyReportResult {
	prrs := []*typepolr.PolicyReportResult{}
	for _, polr := range r.policyReportList.Items {
		for _, result := range polr.Results {
			policy := result.Policy
			if policy == name {
				prrs = append(prrs, &result)
			}
		}
	}
	return prrs
}

func (r *ResultToOscal) loadData(path string, out interface{}) error {
	if err := pkg.LoadYamlFileToK8sTypedObject(r.policyResultsDir+"/"+path, &out); err != nil {
		return err
	}
	return nil
}

func makeProp(name string, value string) policy.Property {
	return policy.Property{
		Name:  name,
		Value: value,
	}
}

func (r *ResultToOscal) GenerateResults() (policy.PVPResult, error) {
	var polList kyvernov1.PolicyList
	if err := r.loadData("/policies.kyverno.io.yaml", &polList); err != nil {
		return policy.PVPResult{}, err
	}

	var cpolList kyvernov1.ClusterPolicyList
	if err := r.loadData("/clusterpolicies.kyverno.io.yaml", &cpolList); err != nil {
		return policy.PVPResult{}, err
	}

	var polrList typepolr.PolicyReportList
	if err := r.loadData("/policyreports.wgpolicyk8s.io.yaml", &polrList); err != nil {
		return policy.PVPResult{}, err
	}
	r.policyReportList = &polrList

	var cpolrList typepolr.ClusterPolicyReportList
	if err := r.loadData("/clusterpolicyreports.wgpolicyk8s.io.yaml", &cpolrList); err != nil {
		return policy.PVPResult{}, err
	}

	var observations []policy.ObservationByCheck
	for _, rule := range r.policy {
		for _, check := range rule.Checks {
			name := check.ID
			prrs := r.retrievePolicyReportResults(name)
			observation := policy.ObservationByCheck{
				Title:       name,
				Description: fmt.Sprintf("Observation of check %s", name),
				Methods:     []string{"TEST-AUTOMATED"},
				Props: []policy.Property{
					makeProp("assessment-rule-id", rule.Rule.ID),
				},
				Subjects: []policy.Subject{},
			}
			for _, prr := range prrs {
				for _, resource := range prr.Subjects {
					gvknsn := fmt.Sprintf("ApiVersion: %s, Kind: %s, Namespace: %s, Name: %s", resource.APIVersion, resource.Kind, resource.Namespace, resource.Name)
					subject := policy.Subject{
						Title:  gvknsn,
						Type:   "resource",
						Result: mapResults(prr.Result),
						Reason: prr.Description,
					}
					observation.Subjects = append(observation.Subjects, subject)
				}
			}
			observations = append(observations, observation)
		}
	}
	result := policy.PVPResult{
		ObservationsByCheck: observations,
	}

	return result, nil
}

func mapResults(result typepolr.PolicyResult) policy.Result {
	switch result {
	case "pass":
		return policy.ResultPass
	case "fail", "warn":
		return policy.ResultFail
	case "error":
		return policy.ResultError
	default:
		return policy.ResultInvalid
	}
}

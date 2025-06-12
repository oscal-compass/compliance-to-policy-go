/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"fmt"
	"slices"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/hashicorp/go-hclog"
	"github.com/oscal-compass/oscal-sdk-go/extensions"

	tp "github.com/oscal-compass/compliance-to-policy-go/v2/framework/template"
)

// ResultsTemplateValues defines values for a plan-based posture report.
type ResultsTemplateValues struct {
	Catalog    string
	Components []tp.Component
}

func CreateResultsValues(
	catalog oscalTypes.Catalog,
	assessmentPlan oscalTypes.AssessmentPlan,
	assessmentResults oscalTypes.AssessmentResults,
	logger hclog.Logger,
) (*ResultsTemplateValues, error) {
	catalogTitle, err := getCatalogTitle(catalog)
	if err != nil {
		return nil, err
	}

	templateValues := &ResultsTemplateValues{
		Catalog: catalogTitle,
	}

	if assessmentPlan.LocalDefinitions == nil || assessmentPlan.LocalDefinitions.Components == nil {
		logger.Warn("assessment plan does not contain components")
		return templateValues, nil
	}

	findings := allFindings(assessmentResults, logger)

	// Attach these to components
	for _, component := range *assessmentPlan.LocalDefinitions.Components {
		tpComp := tp.Component{
			ComponentTitle: component.Title,
		}
		if component.Props == nil {
			continue
		}

		ruleIdsProps := extensions.FindAllProps(*component.Props, extensions.WithName(extensions.RuleIdProp))
		var ruleSet []string
		for _, ruleId := range ruleIdsProps {
			ruleSet = append(ruleSet, ruleId.Value)
		}

		for _, finding := range findings {
			tpFinding := tp.Findings{
				ControlID: finding.ControlID,
			}
			for _, result := range finding.Results {
				// Only add in-scope results to this instance of the finding
				if slices.Contains(ruleSet, result.RuleId) {
					tpFinding.Results = append(tpFinding.Results, result)
				}
			}

			if len(tpFinding.Results) > 0 {
				tpComp.Findings = append(tpComp.Findings, tpFinding)
			}
		}
		templateValues.Components = append(templateValues.Components, tpComp)
	}

	return templateValues, nil
}

// Get the catalog title as the template.md catalog info
func getCatalogTitle(catalog oscalTypes.Catalog) (string, error) {
	if catalog.Metadata.Title != "" {
		return catalog.Metadata.Title, nil
	} else {
		return "", fmt.Errorf("error getting catalog title")
	}
}

// Get controlId info from finding.Target.TargetId
func extractControlId(targetId string) string {
	controlId, _ := strings.CutSuffix(targetId, "_smt")
	return controlId
}

func allFindings(assessmentResults oscalTypes.AssessmentResults, logger hclog.Logger) []tp.Findings {
	var findings []tp.Findings
	observations := make(map[string]oscalTypes.Observation)
	for _, ar := range assessmentResults.Results {
		if ar.Observations == nil {
			continue
		}
		for _, ob := range *ar.Observations {
			if ob.Props == nil || ob.Subjects == nil {
				logger.Debug(fmt.Sprintf("no subjects found for %s", ob.Title))
				continue
			}
			observations[ob.UUID] = ob
		}

		if ar.Findings != nil {
			for _, finding := range *ar.Findings {
				item := tp.Findings{
					ControlID: extractControlId(finding.Target.TargetId),
				}

				if finding.RelatedObservations == nil {
					continue
				}

				for _, relatedObs := range *finding.RelatedObservations {
					ob, found := observations[relatedObs.ObservationUuid]
					if !found {
						logger.Debug(fmt.Sprintf("observation %v not found", relatedObs.ObservationUuid))
						continue
					}

					// Observations with nil Props and Subjects are filtered out when
					// observations are collected.
					ruleId, found := extensions.GetTrestleProp(extensions.AssessmentRuleIdProp, *ob.Props)
					if !found {
						continue
					}
					ruleResult := tp.RuleResult{
						RuleId:   ruleId.Value,
						Subjects: *ob.Subjects,
					}
					item.Results = append(item.Results, ruleResult)
				}
				findings = append(findings, item)
			}
		}
	}
	return findings
}

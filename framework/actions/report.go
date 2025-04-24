/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/defenseunicorns/go-oscal/src/pkg/uuid"
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/rules"
	"github.com/oscal-compass/oscal-sdk-go/transformers"

	"github.com/oscal-compass/compliance-to-policy-go/v2/logging"
	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

// Report action generates an Assessment Results from an Assessment Plan and Context.
func Report(ctx context.Context, inputContext *InputContext, planHref string, plan oscalTypes.AssessmentPlan, results []policy.PVPResult) (*oscalTypes.AssessmentResults, error) {
	log := logging.GetLogger("reporter")
	log.Info(fmt.Sprintf("generating assessments results for plan %s", planHref))

	// for each PVPResult.Observation create an OSCAL Observation
	oscalObservations := make([]oscalTypes.Observation, 0)
	oscalFindings := make([]oscalTypes.Finding, 0)
	store := inputContext.Store()

	// Get all the control mappings based on the assessment activities
	rulesByControls := make(map[string][]string)
	for _, act := range *plan.LocalDefinitions.Activities {
		var controlSet []string
		controls := act.RelatedControls.ControlSelections
		for _, ctr := range controls {
			for _, assess := range *ctr.IncludeControls {
				targetId := fmt.Sprintf("%s_smt", assess.ControlId)
				controlSet = append(controlSet, targetId)
			}
		}
		rulesByControls[act.Title] = controlSet
	}

	// Process into observations
	for _, result := range results {
		for _, observationByCheck := range result.ObservationsByCheck {
			rule, err := store.GetByCheckID(ctx, observationByCheck.CheckID)
			if err != nil {
				if !errors.Is(err, rules.ErrRuleNotFound) {
					return nil, fmt.Errorf("failed to convert observation for check %v: %w", observationByCheck.CheckID, err)
				} else {
					log.Warn(fmt.Sprintf("skipping observation for check %v: %v", observationByCheck.CheckID, err))
					continue
				}
			}
			obs := toOscalObservation(observationByCheck, rule)
			oscalObservations = append(oscalObservations, obs)

			targets, found := rulesByControls[rule.Rule.ID]
			if !found {
				continue
			}

			// if the observation subject result prop is not "pass" then create relevant findings
			if obs.Subjects != nil {
				for _, subject := range *obs.Subjects {
					for _, prop := range *subject.Props {
						if prop.Name == "result" {
							if prop.Value != policy.ResultPass.String() {
								oscalFindings, err = generateFindings(oscalFindings, obs, targets)
								if err != nil {
									return nil, fmt.Errorf("failed to create finding for check: %w", err)
								}
								log.Info(fmt.Sprintf("generated finding for rule %s for subject %s", rule.Rule.ID, subject.Title))
								break
							}
						}
					}
				}
			}

		}
	}

	assessmentResults, err := transformers.AssessmentPlanToAssessmentResults(plan, planHref, oscalObservations...)
	if err != nil {
		return nil, err
	}

	// New assessment results should only have one Assessment Results
	if len(assessmentResults.Results) != 1 {
		return nil, errors.New("bug: assessment results should only have one result")
	}
	assessmentResults.Results[0].Findings = pkg.NilIfEmpty(&oscalFindings)

	return assessmentResults, nil
}

// getFindingForTarget returns an existing finding that matches the targetId if one exists in findings
func getFindingForTarget(findings []oscalTypes.Finding, targetId string) *oscalTypes.Finding {

	for i := range findings {
		if findings[i].Target.TargetId == targetId {
			return &findings[i] // if finding is found, return a pointer to that slice element
		}
	}
	return nil
}

// Generate OSCAL Findings for all non-passing controls in the OSCAL Observation
func generateFindings(findings []oscalTypes.Finding, observation oscalTypes.Observation, targets []string) ([]oscalTypes.Finding, error) {
	for _, targetId := range targets {
		finding := getFindingForTarget(findings, targetId)
		if finding == nil { // if an empty finding was returned, create a new one and append to findings
			newFinding := oscalTypes.Finding{
				UUID: uuid.NewUUID(),
				RelatedObservations: &[]oscalTypes.RelatedObservation{
					{
						ObservationUuid: observation.UUID,
					},
				},
				Target: oscalTypes.FindingTarget{
					TargetId: targetId,
					Type:     "statement-id",
					Status: oscalTypes.ObjectiveStatus{
						State: "not-satisfied",
					},
				},
			}
			findings = append(findings, newFinding)
		} else {
			relObs := oscalTypes.RelatedObservation{
				ObservationUuid: observation.UUID,
			}
			*finding.RelatedObservations = append(*finding.RelatedObservations, relObs) // add new related obs to existing finding for targetId
		}
	}
	return findings, nil
}

// Convert a PVP ObservationByCheck to an OSCAL Observation
func toOscalObservation(observationByCheck policy.ObservationByCheck, ruleSet extensions.RuleSet) oscalTypes.Observation {
	subjects := make([]oscalTypes.SubjectReference, 0)
	for _, subject := range observationByCheck.Subjects {

		props := []oscalTypes.Property{
			{
				Name:  "resource-id",
				Value: subject.ResourceID,
				Ns:    extensions.TrestleNameSpace,
			},
			{
				Name:  "result",
				Value: subject.Result.String(),
				Ns:    extensions.TrestleNameSpace,
			},
			{
				Name:  "evaluated-on",
				Value: subject.EvaluatedOn.String(),
				Ns:    extensions.TrestleNameSpace,
			},
			{
				Name:  "reason",
				Value: subject.Reason,
				Ns:    extensions.TrestleNameSpace,
			},
		}

		s := oscalTypes.SubjectReference{
			SubjectUuid: uuid.NewUUID(),
			Title:       subject.Title,
			Type:        subject.Type,
			Props:       &props,
		}
		subjects = append(subjects, s)
	}

	relevantEvidences := make([]oscalTypes.RelevantEvidence, 0)
	if observationByCheck.RelevantEvidences != nil {
		for _, relEv := range observationByCheck.RelevantEvidences {
			oscalRelEv := oscalTypes.RelevantEvidence{
				Href:        relEv.Href,
				Description: relEv.Description,
			}
			relevantEvidences = append(relevantEvidences, oscalRelEv)
		}
	}

	oscalObservation := oscalTypes.Observation{
		UUID:             uuid.NewUUID(),
		Title:            observationByCheck.Title,
		Description:      observationByCheck.Description,
		Methods:          observationByCheck.Methods,
		Collected:        observationByCheck.Collected,
		Subjects:         pkg.NilIfEmpty(&subjects),
		RelevantEvidence: pkg.NilIfEmpty(&relevantEvidences),
	}

	props := []oscalTypes.Property{
		{
			Name:  "assessment-rule-id",
			Value: ruleSet.Rule.ID,
			Ns:    extensions.TrestleNameSpace,
		},
	}
	oscalObservation.Props = &props

	return oscalObservation
}

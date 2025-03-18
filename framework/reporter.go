/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/defenseunicorns/go-oscal/src/pkg/uuid"
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/hashicorp/go-hclog"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/rules"
	"github.com/oscal-compass/oscal-sdk-go/settings"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/config"
	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

type Reporter struct {
	log        hclog.Logger
	rulesStore rules.Store
}

func NewReporter(cfg *config.C2PConfig) (*Reporter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	rulesStore, _, err := config.ResolveOptions(cfg)
	if err != nil {
		return nil, err
	}

	return &Reporter{
		log:        cfg.Logger,
		rulesStore: rulesStore,
	}, nil
}

type generateOpts struct {
	title string
}

func (g *generateOpts) defaults() {
	g.title = models.SampleRequiredString
}

// GenerateOption defines optional arguments to tune the behavior of GenerateAssessmentResults
type GenerateOption func(opts *generateOpts)

// WithTitle is a GenerateOptions that sets the AssessmentResults title in the metadata
func WithTitle(title string) GenerateOption {
	return func(opts *generateOpts) {
		opts.title = title
	}
}

// getFindingForTarget returns an existing finding that matches the targetId if one exists in findings
func (r *Reporter) getFindingForTarget(findings []oscalTypes.Finding, targetId string) *oscalTypes.Finding {

	for i := range findings {
		if findings[i].Target.TargetId == targetId {
			return &findings[i] // if finding is found, return a pointer to that slice element
		}
	}
	return nil
}

// Generate OSCAL Findings for all non-passing controls in the OSCAL Observation
func (r *Reporter) generateFindings(findings []oscalTypes.Finding, observation oscalTypes.Observation, ruleSet extensions.RuleSet, implementationSettings settings.ImplementationSettings) ([]oscalTypes.Finding, error) {
	applicableControls, err := implementationSettings.ApplicableControls(ruleSet.Rule.ID)
	if err != nil {
		return findings, err
	}

	for _, control := range applicableControls {

		targetId := fmt.Sprintf("%s_smt", control.ControlId)

		finding := r.getFindingForTarget(findings, targetId)

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

// findControls finds all controls from the implementation settings
func (r *Reporter) findControls(implementationSettings settings.ImplementationSettings) oscalTypes.ReviewedControls {

	includeControls := implementationSettings.AllControls()

	assessedControls := []oscalTypes.AssessedControls{
		{
			IncludeControls: &includeControls,
		},
	}

	reviewedConrols := oscalTypes.ReviewedControls{
		ControlSelections: assessedControls,
	}
	return reviewedConrols

}

// Convert a PVP ObservationByCheck to an OSCAL Observation
func (r *Reporter) toOscalObservation(observationByCheck policy.ObservationByCheck, ruleSet extensions.RuleSet) oscalTypes.Observation {
	subjects := make([]oscalTypes.SubjectReference, 0)
	for _, subject := range observationByCheck.Subjects {

		props := []oscalTypes.Property{
			{
				Name:  "resource-id",
				Value: subject.ResourceID,
			},
			{
				Name:  "result",
				Value: subject.Result.String(),
			},
			{
				Name:  "evaluated-on",
				Value: subject.EvaluatedOn.String(),
			},
			{
				Name:  "reason",
				Value: subject.Reason,
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
		},
	}
	oscalObservation.Props = &props

	return oscalObservation
}

// GenerateAssessmentResults converts PVPResults to OSCAL AsessmentResults
func (r *Reporter) GenerateAssessmentResults(ctx context.Context, planHref string, implementationSettings *settings.ImplementationSettings, results []policy.PVPResult, opts ...GenerateOption) (oscalTypes.AssessmentResults, error) {

	options := generateOpts{}
	options.defaults()
	for _, opt := range opts {
		opt(&options)
	}

	r.log.Info(fmt.Sprintf("generating assessments results for plan %s", planHref))

	importAp := oscalTypes.ImportAp{
		Href: planHref,
	}

	metadata := models.NewSampleMetadata()
	metadata.Title = options.title

	assessmentResults := oscalTypes.AssessmentResults{
		UUID:     uuid.NewUUID(),
		ImportAp: importAp,
		Metadata: metadata,
	}

	// for each PVPResult.Observation create an OSCAL Observation
	oscalObservations := make([]oscalTypes.Observation, 0)
	oscalFindings := make([]oscalTypes.Finding, 0)

	for _, result := range results {

		for _, observationByCheck := range result.ObservationsByCheck {
			rule, err := r.rulesStore.GetByCheckID(ctx, observationByCheck.CheckID)
			if err != nil {
				if !errors.Is(err, rules.ErrRuleNotFound) {
					return assessmentResults, fmt.Errorf("failed to convert observation for check %v: %w", observationByCheck.CheckID, err)
				} else {
					r.log.Warn(fmt.Sprintf("skipping observation for check %v: %v", observationByCheck.CheckID, err))
					continue
				}
			}

			obs := r.toOscalObservation(observationByCheck, rule)

			// if the observation subject result prop is not "pass" then create relevant findings
			if obs.Subjects != nil {
				for _, subject := range *obs.Subjects {
					for _, prop := range *subject.Props {
						if prop.Name == "result" {
							if prop.Value != policy.ResultPass.String() {
								oscalFindings, err = r.generateFindings(oscalFindings, obs, rule, *implementationSettings)
								if err != nil {
									return assessmentResults, fmt.Errorf("failed to create finding for check: %w", err)
								}
								r.log.Info(fmt.Sprintf("generated finding for rule %s", rule.Rule.ID))

							}
						}
					}
				}
			}

			oscalObservations = append(oscalObservations, obs)
		}

	}
	reviewedControls := r.findControls(*implementationSettings)

	oscalResult := oscalTypes.Result{
		UUID:             uuid.NewUUID(),
		Title:            "Automated Assessment Result",
		Description:      "Assessment Results Automatically Genererated from PVP Results",
		Start:            time.Now(),
		ReviewedControls: reviewedControls,
		Observations:     &oscalObservations,
	}

	if len(oscalFindings) > 0 {
		oscalResult.Findings = &oscalFindings
	}

	assessmentResults.Results = []oscalTypes.Result{
		oscalResult,
	}

	return assessmentResults, nil
}

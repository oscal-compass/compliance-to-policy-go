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
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/models"
	"github.com/oscal-compass/oscal-sdk-go/rules"
	"github.com/oscal-compass/oscal-sdk-go/settings"
)

const (
	defaultVersion = "0.1.0"
	defaultTitle   = "Automated Assessment Results"
)

type Reporter struct {
	log   hclog.Logger
	store rules.Store
}

func NewReporter(log hclog.Logger, store rules.Store) *Reporter {
	return &Reporter{
		store: store,
	}
}

type generateOpts struct {
	title string
}

func (g *generateOpts) defaults() {
	g.title = models.DefaultRequiredString
}

// GenerateOption defineoptions to tune the behavior of
// GenerateAssessmentResults.
type GenerateOption func(opts *generateOpts)

// WithTitle is a GenerateOptions that sets the AssessmentResults title
// in the metadata.
func WithTitle(title string) GenerateOption {
	return func(opts *generateOpts) {
		opts.title = title
	}
}

func (r *Reporter) findControls(implementationSettings settings.ImplementationSettings) oscalTypes.ReviewedControls {

	includeControls := []oscalTypes.AssessedControlsSelectControlById{}

	for _, controlId := range implementationSettings.AllControls() {
		selectedControlById := oscalTypes.AssessedControlsSelectControlById{
			ControlId: controlId,
		}
		includeControls = append(includeControls, selectedControlById)
	}

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

func (r *Reporter) toOscalObservation(observationByCheck policy.ObservationByCheck, ruleSet extensions.RuleSet) (oscalTypes.Observation, error) {

	oscalObservation := oscalTypes.Observation{
		UUID:        uuid.NewUUID(),
		Title:       observationByCheck.Title,
		Description: observationByCheck.Description,
		Methods:     observationByCheck.Methods,
		Collected:   observationByCheck.Collected,
	}

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
	oscalObservation.Subjects = &subjects

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
	if len(relevantEvidences) > 0 {
		oscalObservation.RelevantEvidence = &relevantEvidences
	}

	props := []oscalTypes.Property{
		{
			Name:  "assessment-rule-id",
			Value: ruleSet.Rule.ID,
		},
	}
	oscalObservation.Props = &props

	return oscalObservation, nil
}

// Convert PVPResults to OSCAL AsessmentResults
func (r *Reporter) GenerateAssessmentResults(ctx context.Context, planHref string, implementationSettings *settings.ImplementationSettings, results []*policy.PVPResult, opts ...GenerateOption) (oscalTypes.AssessmentResults, error) {

	options := generateOpts{}
	options.defaults()
	for _, opt := range opts {
		opt(&options)
	}

	importAp := oscalTypes.ImportAp{
		Href: planHref,
	}

	metadata := oscalTypes.Metadata{
		Title:        options.title,
		LastModified: time.Now(),
		Version:      defaultVersion,
		OscalVersion: models.OSCALVersion,
	}

	assessmentResults := oscalTypes.AssessmentResults{
		UUID:     uuid.NewUUID(),
		ImportAp: importAp,
		Metadata: metadata,
	}

	// for each PVPResult.Observation create an OSCAL Observation
	oscalObservations := make([]oscalTypes.Observation, 0)
	for _, result := range results {

		for _, observationByCheck := range result.ObservationsByCheck {
			rule, err := r.store.GetByCheckID(ctx, observationByCheck.CheckID)
			if err != nil {
				if !errors.Is(err, rules.ErrRuleNotFound) {
					return assessmentResults, fmt.Errorf("failed to convert observation for check: %w", err)
				}
			}

			obs, err := r.toOscalObservation(observationByCheck, rule)
			if err != nil {
				return assessmentResults, fmt.Errorf("failed to convert observation for check: %w", err)
			}
			oscalObservations = append(oscalObservations, obs)
		}
	}

	reviewedConrols := r.findControls(*implementationSettings)

	oscalResults := []oscalTypes.Result{
		{
			UUID:             uuid.NewUUID(),
			Title:            "Automated Assessment Result",
			Description:      "Assessment Results Automatically Genererated from PVP Results",
			Start:            time.Now(),
			ReviewedControls: reviewedConrols,
			Observations:     &oscalObservations,
		},
	}
	assessmentResults.Results = oscalResults

	return assessmentResults, nil

}

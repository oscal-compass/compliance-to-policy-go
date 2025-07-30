/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

import (
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/oscal-compass/compliance-to-policy-go/v2/api/proto"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

// PolicyToProto transforms a plugin Policy to a protobuf PolicyRequest.
func PolicyToProto(p policy.Policy) []*proto.Rule {
	var rules []*proto.Rule
	for _, rs := range p {
		var parameters []*proto.Parameter
		for _, prm := range rs.Rule.Parameters {
			protoPrm := &proto.Parameter{
				Name:          prm.ID,
				Description:   prm.Description,
				SelectedValue: prm.Value,
			}
			parameters = append(parameters, protoPrm)
		}

		var checks []*proto.Check
		for _, ch := range rs.Checks {
			check := &proto.Check{
				Name:        ch.ID,
				Description: ch.Description,
			}
			checks = append(checks, check)
		}
		ruleSet := &proto.Rule{
			Name:        rs.Rule.ID,
			Description: rs.Rule.Description,
			Checks:      checks,
			Parameters:  parameters,
		}
		rules = append(rules, ruleSet)
	}
	return rules
}

// NewPolicyFromProto transforms protobuf PolicyRequest into a plugin Policy.
func NewPolicyFromProto(rules []*proto.Rule) policy.Policy {
	var p policy.Policy

	for _, r := range rules {
		var parameters []extensions.Parameter
		for _, prm := range r.Parameters {
			parameter := extensions.Parameter{
				ID:          prm.Name,
				Description: prm.Description,
				Value:       prm.SelectedValue,
			}
			parameters = append(parameters, parameter)
		}
		var checks []extensions.Check
		for _, ch := range r.Checks {
			check := extensions.Check{
				ID:          ch.Name,
				Description: ch.Description,
			}
			checks = append(checks, check)
		}

		rule := extensions.RuleSet{
			Rule: extensions.Rule{
				ID:          r.Name,
				Description: r.Description,
				Parameters:  parameters,
			},
			Checks: checks,
		}

		p = append(p, rule)
	}
	return p
}

var protoByResult = map[policy.Result]proto.Result{
	policy.ResultPass:    proto.Result_RESULT_PASS,
	policy.ResultInvalid: proto.Result_RESULT_UNSPECIFIED,
	policy.ResultError:   proto.Result_RESULT_ERROR,
	policy.ResultWarning: proto.Result_RESULT_WARNING,
	policy.ResultFail:    proto.Result_RESULT_FAILURE,
}

var resultByProto = map[proto.Result]policy.Result{
	proto.Result_RESULT_UNSPECIFIED: policy.ResultInvalid,
	proto.Result_RESULT_ERROR:       policy.ResultError,
	proto.Result_RESULT_WARNING:     policy.ResultWarning,
	proto.Result_RESULT_PASS:        policy.ResultPass,
	proto.Result_RESULT_FAILURE:     policy.ResultFail,
}

func NewResultFromProto(pb *proto.PVPResult) policy.PVPResult {
	result := policy.PVPResult{}

	for _, o := range pb.Observations {
		observation := policy.ObservationByCheck{
			Title:       o.Name,
			Description: o.Description,
			Methods:     o.Methods,
			Collected:   o.CollectedAt.AsTime(),
			CheckID:     o.CheckId,
		}
		var links []policy.Link
		for _, ref := range o.EvidenceRefs {
			link := policy.Link{Description: ref.Description, Href: ref.Href}
			links = append(links, link)
		}
		observation.RelevantEvidences = links

		var subjects []policy.Subject
		for _, s := range o.Subjects {
			subject := policy.Subject{
				Title:       s.Title,
				Type:        s.Type,
				ResourceID:  s.ResourceId,
				Result:      resultByProto[s.Result],
				EvaluatedOn: s.EvaluatedOn.AsTime(),
				Reason:      s.Reason,
			}
			var subjectProps []policy.Property
			for _, sp := range s.Props {
				subjectProp := policy.Property{Name: sp.Name, Value: sp.Value}
				subjectProps = append(subjectProps, subjectProp)
			}
			subject.Props = subjectProps
			subjects = append(subjects, subject)
		}
		observation.Subjects = subjects

		var props []policy.Property
		for _, p := range o.Props {
			prop := policy.Property{Name: p.Name, Value: p.Value}
			props = append(props, prop)
		}
		observation.Props = props
		result.ObservationsByCheck = append(result.ObservationsByCheck, observation)
	}

	for _, l := range pb.Links {
		link := policy.Link{Description: l.Description, Href: l.Href}
		result.Links = append(result.Links, link)
	}
	return result
}

func ResultsToProto(result policy.PVPResult) *proto.PVPResult {
	pvpResult := &proto.PVPResult{}

	for _, o := range result.ObservationsByCheck {
		observation := &proto.ObservationByCheck{
			Name:        o.Title,
			Description: o.Description,
			CheckId:     o.CheckID,
			Methods:     o.Methods,
			CollectedAt: timestamppb.New(o.Collected),
		}
		var subjects []*proto.Subject
		for _, s := range o.Subjects {
			subject := &proto.Subject{
				Title:       s.Title,
				Type:        s.Type,
				ResourceId:  s.ResourceID,
				Result:      protoByResult[s.Result],
				EvaluatedOn: timestamppb.New(s.EvaluatedOn),
				Reason:      s.Reason,
			}
			var subjectProps []*proto.Property
			for _, sp := range s.Props {
				subjectProp := &proto.Property{Name: sp.Name, Value: sp.Value}
				subjectProps = append(subjectProps, subjectProp)
			}
			subject.Props = subjectProps
			subjects = append(subjects, subject)
		}
		var evidences []*proto.Link
		for _, evidence := range o.RelevantEvidences {
			link := &proto.Link{Description: evidence.Description, Href: evidence.Href}
			evidences = append(evidences, link)
		}
		var props []*proto.Property
		for _, p := range o.Props {
			prop := &proto.Property{Name: p.Name, Value: p.Value}
			props = append(props, prop)
		}
		observation.EvidenceRefs = evidences
		observation.Subjects = subjects
		observation.Props = props
		pvpResult.Observations = append(pvpResult.Observations, observation)
	}

	for _, l := range result.Links {
		link := &proto.Link{
			Description: l.Description,
			Href:        l.Href,
		}
		pvpResult.Links = append(pvpResult.Links, link)

	}
	return pvpResult
}

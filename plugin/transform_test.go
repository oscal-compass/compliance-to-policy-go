/*
 Copyright 2024 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package plugin

import (
	"testing"
	"time"

	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/oscal-compass/compliance-to-policy-go/v2/api/proto"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

var testTimeString, _ = time.Parse("00:00:00", "12:00:00")

var testPolicy = policy.Policy{
	extensions.RuleSet{
		Rule: extensions.Rule{
			ID:          "test-rule-1",
			Description: "test rule 1",
			Parameters: []extensions.Parameter{
				{
					ID:          "test-param-1",
					Description: "test param 1",
					Value:       "test param value",
				},
			},
		},
		Checks: []extensions.Check{
			{
				ID:          "test-check-1",
				Description: "test check 1",
			},
		},
	},
}

var testPolicyRequest = &proto.PolicyRequest{
	Rule: []*proto.Rule{
		{
			Name:        "test-rule-1",
			Description: "test rule 1",
			Checks: []*proto.Check{
				{
					Name:        "test-check-1",
					Description: "test check 1",
				},
			},
			Parameters: []*proto.Parameter{
				{
					Name:          "test-param-1",
					Description:   "test param 1",
					SelectedValue: "test param value",
				},
			},
		},
	},
}

var testProtoPvpResult = &proto.PVPResult{
	Observations: []*proto.ObservationByCheck{
		{
			Name:        "test-obs-1",
			Description: "test obs 1",
			CheckId:     "test-check-1",
			Methods:     []string{"method-1", "method-2"},
			CollectedAt: timestamppb.New(testTimeString),
			Subjects: []*proto.Subject{
				{
					Title:       "test-subject-1",
					Type:        "test",
					ResourceId:  "test-resource-1",
					Result:      proto.Result_RESULT_PASS,
					EvaluatedOn: timestamppb.New(testTimeString),
					Reason:      "test reason",
					Props: []*proto.Property{
						{
							Name:  "test-subject-prop-1",
							Value: "test-subject-value-1",
						},
					},
				},
			},
			EvidenceRefs: []*proto.Link{
				{
					Description: "test-evidence-ref-1",
					Href:        "https://test-evidence-ref-1",
				},
			},
			Props: []*proto.Property{
				{
					Name:  "test-prop-1",
					Value: "test value 1",
				},
			},
		},
	},
	Links: []*proto.Link{
		{
			Description: "test-link-1",
			Href:        "https://test-link-1",
		},
	},
}

var testPolicyPvpResult = policy.PVPResult{
	ObservationsByCheck: []policy.ObservationByCheck{
		{
			Title:       "test-obs-1",
			Description: "test obs 1",
			CheckID:     "test-check-1",
			Methods:     []string{"method-1", "method-2"},
			Subjects: []policy.Subject{
				{
					Title:       "test-subject-1",
					Type:        "test",
					ResourceID:  "test-resource-1",
					Result:      policy.ResultPass,
					EvaluatedOn: timestamppb.New(testTimeString).AsTime(),
					Reason:      "test reason",
					Props: []policy.Property{
						{
							Name:  "test-subject-prop-1",
							Value: "test-subject-value-1",
						},
					},
				},
			},
			Collected: timestamppb.New(testTimeString).AsTime(),
			RelevantEvidences: []policy.Link{
				{
					Description: "test-evidence-ref-1",
					Href:        "https://test-evidence-ref-1",
				},
			},
			Props: []policy.Property{
				{
					Name:  "test-prop-1",
					Value: "test value 1",
				},
			},
		},
	},
	Links: []policy.Link{
		{
			Description: "test-link-1",
			Href:        "https://test-link-1",
		},
	},
}

func TestPolicyToProto(t *testing.T) {
	output := PolicyToProto(testPolicy)
	require.Equal(t, testPolicyRequest, output)
}

func TestNewPolicyFromProto(t *testing.T) {
	output := NewPolicyFromProto(testPolicyRequest)
	require.Equal(t, testPolicy, output)
}

func TestResultToProto(t *testing.T) {
	output := ResultsToProto(testPolicyPvpResult)
	require.Equal(t, testProtoPvpResult, output)
}

func TestNewResultFromProto(t *testing.T) {
	output := NewResultFromProto(testProtoPvpResult)
	require.Equal(t, testPolicyPvpResult, output)
}

/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package actions

import (
	"context"
	"sort"
	"testing"

	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/settings"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/oscal-compass/compliance-to-policy-go/v2/plugin"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

var (
	expectedCertFileRule = extensions.RuleSet{
		Rule: extensions.Rule{
			ID:          "etcd_cert_file",
			Description: "Ensure that the --cert-file argument is set as appropriate",
		},
		Checks: []extensions.Check{
			{
				ID:          "etcd_cert_file",
				Description: "Check that the --cert-file argument is set as appropriate",
			},
		},
	}
	expectedKeyFileRule = extensions.RuleSet{
		Rule: extensions.Rule{
			ID:          "etcd_key_file",
			Description: "Ensure that the --key-file argument is set as appropriate",
			Parameters: []extensions.Parameter{
				{
					ID:          "file_name",
					Description: "A parameter for a file name",
				},
			},
		},
		Checks: []extensions.Check{
			{
				ID:          "etcd_key_file",
				Description: "Check that the --key-file argument is set as appropriate",
			},
		},
	}
)

func TestAggregateResults(t *testing.T) {
	inputContext := inputContextHelper(t)
	wantResults := policy.PVPResult{
		ObservationsByCheck: []policy.ObservationByCheck{
			{
				Title:       "Example",
				Description: "Example description",
				CheckID:     "test-check",
			},
		},
	}

	updatedParam := extensions.Parameter{
		ID:          "file_name",
		Description: "A parameter for a file name",
		Value:       "my_file",
	}

	updatedKeyFileRule := expectedKeyFileRule
	updatedKeyFileRule.Rule.Parameters[0] = updatedParam

	// Create pluginSet
	providerTestObj := new(policyProvider)
	providerTestObj.On("GetResults", policy.Policy{updatedKeyFileRule}).Return(wantResults, nil)
	pluginSet := map[plugin.ID]policy.Provider{
		"mypvpvalidator": providerTestObj,
	}

	testSettings := settings.NewSettings(map[string]struct{}{"etcd_key_file": {}}, map[string]string{"file_name": "my_file"})
	inputContext.Settings = testSettings

	gotResults, err := AggregateResults(context.TODO(), inputContext, pluginSet)
	require.NoError(t, err)
	providerTestObj.AssertExpectations(t)
	require.Len(t, gotResults, 1)
}

// policyProvider is a mocked implementation of policy.Provider.
type policyProvider struct {
	mock.Mock
}

func (p *policyProvider) Configure(option map[string]string) error {
	args := p.Called(option)
	return args.Error(0)
}

func (p *policyProvider) Generate(policyRules policy.Policy) error {
	sort.SliceStable(policyRules, func(i, j int) bool {
		return policyRules[i].Rule.ID > policyRules[j].Rule.ID
	})
	args := p.Called(policyRules)
	return args.Error(0)
}

func (p *policyProvider) GetResults(policyRules policy.Policy) (policy.PVPResult, error) {
	sort.SliceStable(policyRules, func(i, j int) bool {
		return policyRules[i].Rule.ID > policyRules[j].Rule.ID
	})
	args := p.Called(policyRules)
	return args.Get(0).(policy.PVPResult), args.Error(1)
}

/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package actions

import (
	"context"
	"testing"

	"github.com/oscal-compass/oscal-sdk-go/settings"
	"github.com/stretchr/testify/require"

	"github.com/oscal-compass/compliance-to-policy-go/v2/plugin"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

func TestGeneratePolicy(t *testing.T) {
	inputContext := inputContextHelper(t)

	// Create pluginSet
	providerTestObj := new(policyProvider)
	providerTestObj.On("Generate", policy.Policy{expectedCertFileRule}).Return(nil)
	pluginSet := map[plugin.ID]policy.Provider{
		"mypvpvalidator": providerTestObj,
	}

	testSettings := settings.NewSettings(map[string]struct{}{"etcd_cert_file": {}}, map[string]string{})
	inputContext.Settings = testSettings

	err := GeneratePolicy(context.TODO(), inputContext, pluginSet)
	require.NoError(t, err)
	providerTestObj.AssertExpectations(t)
}

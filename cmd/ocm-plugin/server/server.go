/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"errors"
	"os"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/hashicorp/go-hclog"

	"github.com/oscal-compass/compliance-to-policy-go/v2/logging"
	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

var (
	_      policy.Provider = (*Plugin)(nil)
	logger hclog.Logger    = logging.NewPluginLogger()
)

func Logger() hclog.Logger {
	return logger
}

type Plugin struct {
	config             Config
	policyGeneratorDir string
}

func NewPlugin() *Plugin {
	return &Plugin{
		config: Config{},
	}
}

func (p *Plugin) Configure(m map[string]string) error {
	if err := mapstructure.Decode(m, &p.config); err != nil {
		return errors.New("error decoding configuration")
	}
	return p.config.Validate()
}

func (p *Plugin) Generate(pl policy.Policy) error {
	tmpdir := pkg.NewTempDirectory(p.config.TempDir)
	composer := NewComposerByTempDirectory(p.config.PoliciesDir, tmpdir)
	if err := composer.ComposeByPolicies(pl, p.config); err != nil {
		return err
	}
	policySet, err := composer.GeneratePolicySet()
	if err != nil {
		return err
	}

	for _, resource := range (*policySet).Resources() {
		name := resource.GetName()
		kind := resource.GetKind()
		namespace := resource.GetNamespace()
		yamlByte, err := resource.AsYAML()
		if err != nil {
			return err
		}
		fnamesTokens := []string{kind, namespace, name}
		fname := strings.Join(fnamesTokens, ".") + ".yaml"
		if err := os.WriteFile(p.config.OutputDir+"/"+fname, yamlByte, os.ModePerm); err != nil {
			return err
		}
	}

	if p.policyGeneratorDir != "" {
		if err := composer.CopyAllTo(p.policyGeneratorDir); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plugin) GetResults(pl policy.Policy) (policy.PVPResult, error) {
	results := NewResultToOscal(pl, p.config.PolicyResultsDir, p.config.Namespace, p.config.PolicySetName)
	return results.GenerateResults()
}

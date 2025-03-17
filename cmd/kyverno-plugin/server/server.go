/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"errors"

	"github.com/go-viper/mapstructure/v2"
	"github.com/hashicorp/go-hclog"

	"github.com/oscal-compass/compliance-to-policy-go/v2/internal/logging"
	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

var (
	_      policy.Provider = (*Plugin)(nil)
	logger                 = logging.GetLogger("kyverno-server")
)

type Plugin struct {
	config Config
	logger hclog.Logger
}

func NewPlugin() *Plugin {
	return &Plugin{
		logger: logger,
	}
}

func (p *Plugin) Configure(m map[string]string) error {
	if err := mapstructure.Decode(m, &p.config); err != nil {
		return errors.New("error decoding configuration")
	}
	return p.config.Validate()
}

func (p *Plugin) Generate(pl policy.Policy) error {
	logger.Debug(p.config.PoliciesDir)
	tmpdir := pkg.NewTempDirectory(p.config.TempDir)
	composer := NewOscal2Policy(p.config.PoliciesDir, tmpdir)
	if err := composer.Generate(pl); err != nil {
		return err
	}

	if p.config.OutputDir != "" {
		if err := composer.CopyAllTo(p.config.OutputDir); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plugin) GetResults(pl policy.Policy) (policy.PVPResult, error) {
	results := NewResultToOscal(pl, p.config.PolicyResultsDir)
	return results.GenerateResults()
}

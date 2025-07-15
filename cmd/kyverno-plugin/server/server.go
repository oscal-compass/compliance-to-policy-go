/*
 Copyright 2025 The OSCAL Compass Authors
 SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"errors"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/hashicorp/go-hclog"

	"github.com/oscal-compass/compliance-to-policy-go/v2/internal/utils"
	"github.com/oscal-compass/compliance-to-policy-go/v2/logging"
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
	config Config
}

func NewPlugin() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Configure(m map[string]string) error {
	if err := mapstructure.Decode(m, &p.config); err != nil {
		return errors.New("error decoding configuration")
	}
	return p.config.Validate()
}

func (p *Plugin) Generate(pl policy.Policy) error {
	logger.Debug(fmt.Sprintf("Using resources from %s", p.config.PoliciesDir))
	tmpdir := utils.NewTempDirectory(p.config.TempDir)
	composer := NewOscal2Policy(p.config.PoliciesDir, tmpdir)
	if err := composer.Generate(pl); err != nil {
		return err
	}

	if p.config.OutputDir != "" {
		if err := composer.CopyAllTo(p.config.OutputDir); err != nil {
			return err
		}
		logger.Debug(fmt.Sprintf("Copied outputs to %s", p.config.OutputDir))
	}
	return nil
}

func (p *Plugin) GetResults(pl policy.Policy) (policy.PVPResult, error) {
	results := NewResultToOscal(pl, p.config.PolicyResultsDir)
	return results.GenerateResults()
}

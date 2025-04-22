/*
Copyright 2023 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"bytes"
	"embed"
	"html/template"
	"os"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/hashicorp/go-hclog"
)

//go:embed template/*.md
var embeddedResources embed.FS

type Oscal2Posture struct {
	logger            hclog.Logger
	assessmentResults *oscalTypes.AssessmentResults
	catalog           *oscalTypes.Catalog
	compDef           *oscalTypes.ComponentDefinition
	templateFile      *string
}

func NewOscal2Posture(assessmentResults *oscalTypes.AssessmentResults, catalog *oscalTypes.Catalog, compDef *oscalTypes.ComponentDefinition, logger hclog.Logger) *Oscal2Posture {
	return &Oscal2Posture{
		assessmentResults: assessmentResults,
		catalog:           catalog,
		compDef:           compDef,
		logger:            logger,
	}
}

func (r *Oscal2Posture) SetTemplateFile(templateFile string) {
	r.templateFile = &templateFile
}

func (r *Oscal2Posture) Generate() ([]byte, error) {
	var templateData []byte
	var err error
	if r.templateFile == nil {
		templateData, err = embeddedResources.ReadFile("template/posture.md")
	} else {
		templateData, err = os.ReadFile(*r.templateFile)
	}
	if err != nil {
		return nil, err
	}

	funcmap := template.FuncMap{
		"newline_with_indent": func(text string, indent int) string {
			newText := strings.ReplaceAll(text, "\n", "\n"+strings.Repeat(" ", indent))
			return newText
		},
	}

	templateString := string(templateData)
	tmpl := template.New("report.md")
	tmpl.Funcs(funcmap)
	tmpl, err = tmpl.Parse(templateString)
	if err != nil {
		return nil, err
	}
	templateValue, err := CreateComponentValues(r.catalog, r.compDef, r.assessmentResults, r.logger)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buffer, templateValue)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

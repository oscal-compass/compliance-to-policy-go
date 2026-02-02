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

package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/oscal-compass/compliance-to-policy-go/v2/internal/tables/resources"
	"github.com/oscal-compass/compliance-to-policy-go/v2/internal/types/policycomposition"
	"github.com/oscal-compass/compliance-to-policy-go/v2/internal/utils"
)

func main() {
	var resourceTableFile, filter, format string
	flag.StringVar(&resourceTableFile, "resource-table-file", utils.PathFromInternalDirectory("../out/resources.csv"), "path to resource table file (csv)")
	flag.StringVar(&filter, "query-params", "", "query-param (e.g. --filter=\"source=xxx&category=yyy\"")
	flag.StringVar(&format, "format", "table", "output format (e.g. available formats are table, yaml")
	flag.Parse()

	u, err := url.Parse("https://abc?" + filter)
	if err != nil {
		panic(err)
	}
	println(u)

	cleanedPath := filepath.Clean(resourceTableFile)
	f, err := os.Open(cleanedPath)
	if err != nil {
		panic(err)
	}

	qvs := u.Query()
	filterFunc := func(row resources.Row) bool {
		conds := []bool{}
		for k, vs := range qvs {
			for _, v := range vs {
				conds = append(conds, row.Get(k) == v)
			}
		}
		cond := true
		for _, c := range conds {
			cond = cond && c
		}
		return cond
	}

	t := resources.FromCsv(f)
	filtered := t.Filter(filterFunc)
	if format == "table" {
		err = filtered.Print()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
			os.Exit(1)
		}
	} else if format == "yaml" {
		resources := []policycomposition.Resource{}
		rows := filtered.List()
		for _, row := range rows {
			resource := policycomposition.Resource{
				ApiVersion: row.ApiVersion,
				Kind:       row.Kind,
				Name:       row.Name,
			}
			resources = append(resources, resource)
		}
		policyComposition := policycomposition.PolicyComposition{
			ApiVersion: "github.ibm.com/poloicy-collection-plus/v1",
			Kind:       "PolicyComposition",
			Metadata: policycomposition.Metadata{
				Name: "policy-composition",
			},
			Compliance: policycomposition.Compliance{
				Standard: rows[0].Standard,
				Category: rows[0].Category,
				Control:  rows[0].Control,
			},
			Resources: resources,
		}
		writer := bufio.NewWriter(os.Stdout)
		yamlByte, err := yaml.Marshal(policyComposition)
		if err != nil {
			panic(err)
		}
		count, err := writer.WriteString(string(yamlByte))
		if err != nil {
			panic(err)
		}
		_ = count
		_ = writer.Flush()
	} else {
		err = t.Filter(filterFunc).Print()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error printing table: %v\n", err)
			os.Exit(1)
		}
	}
}

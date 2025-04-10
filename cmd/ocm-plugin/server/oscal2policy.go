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

package server

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	cp "github.com/otiai10/copy"
	"sigs.k8s.io/kustomize/api/resmap"
	typekustomize "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"

	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg"
	"github.com/oscal-compass/compliance-to-policy-go/v2/pkg/policygenerator"
	pgtype "github.com/oscal-compass/compliance-to-policy-go/v2/pkg/types/policygenerator"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"
)

var DummyNamespace string = "dummy-namespace-c2p"

type Composer struct {
	policiesDir string
	tempDir     pkg.TempDirectory
}

func NewComposer(policiesDir string, tempDir string) *Composer {
	return NewComposerByTempDirectory(policiesDir, pkg.NewTempDirectory(tempDir))
}

func NewComposerByTempDirectory(policiesDir string, tempDir pkg.TempDirectory) *Composer {
	return &Composer{
		policiesDir: policiesDir,
		tempDir:     tempDir,
	}
}

func (c *Composer) GetPoliciesDir() string {
	return c.policiesDir
}

func (c *Composer) ComposeByPolicies(pl policy.Policy, config Config) error {
	return c.Compose(pl, config)
}

func (c *Composer) Compose(pl policy.Policy, config Config) error {
	if config.clusterSelectors == nil {
		config.clusterSelectors = map[string]string{"env": "dev"}
	}

	logger.Info("Start composing policySets")
	parameters := map[string]string{}
	policyConfigMap := map[string]pgtype.PolicyConfig{}
	policySets := []pgtype.PolicySetConfig{}
	policySetPatches := []typekustomize.Patch{}

	logger.Info("Start generating policy")
	for idx, ruleObject := range pl {
		policyListPerControlImple := []string{}
		if ruleObject.Rule.Parameter != nil {
			parameters[ruleObject.Rule.Parameter.ID] = ruleObject.Rule.Parameter.Value
		}
		for _, check := range ruleObject.Checks {
			policyId := check.ID
			sourceDir := fmt.Sprintf("%s/%s", c.policiesDir, check.ID)
			destDir := fmt.Sprintf("%s/%s", c.tempDir.GetTempDir(), policyId)
			err := cp.Copy(sourceDir, destDir)
			if err != nil {
				return err
			}

			policyGeneratorManifestPath := destDir + "/policy-generator.yaml"
			var policyGeneratorManifest pgtype.PolicyGenerator
			if err := pkg.LoadYamlFileToObject(policyGeneratorManifestPath, &policyGeneratorManifest); err != nil {
				return err
			}
			policyGeneratorManifest.PolicyDefaults.Namespace = config.Namespace
			policyGeneratorManifest.PolicyDefaults.PolicyOptions.Placement.ClusterSelectors = config.clusterSelectors
			if err := pkg.WriteObjToYamlFileByGoYaml(policyGeneratorManifestPath, policyGeneratorManifest); err != nil {
				return err
			}
			// For policySet
			policyListPerControlImple = appendUnique(policyListPerControlImple, policyId)
			policyConfig, ok := policyConfigMap[policyId]
			if ok {
				policyConfig.Standards = appendUnique(policyConfig.Standards, policyGeneratorManifest.PolicyDefaults.Standards...)
				policyConfig.Categories = appendUnique(policyConfig.Categories, policyGeneratorManifest.PolicyDefaults.Categories...)
				policyConfig.Controls = appendUnique(policyConfig.Controls, policyGeneratorManifest.PolicyDefaults.Controls...)
				policyConfigMap[policyId] = policyConfig
			} else {
				policyConfig := policyGeneratorManifest.Policies[0]
				policyConfig.Standards = policyGeneratorManifest.PolicyDefaults.Standards
				policyConfig.Categories = policyGeneratorManifest.PolicyDefaults.Categories
				policyConfig.Controls = policyGeneratorManifest.PolicyDefaults.Controls
				for idx, manifest := range policyConfig.Manifests {
					policyConfig.Manifests[idx].Path = strings.Replace(manifest.Path, "./", fmt.Sprintf("./%s/", policyId), 1)
				}
				policyConfigMap[policyId] = policyConfig
			}
		}

		suffix := ""
		if idx > 0 {
			suffix = fmt.Sprintf("-%d", idx)
		}
		policySetConfig := pgtype.PolicySetConfig{
			Name:     toDNSCompliant(config.PolicySetName + suffix),
			Policies: policyListPerControlImple,
		}
		policySets = append(policySets, policySetConfig)
		policySetPatch := typekustomize.Patch{
			Target: &typekustomize.Selector{
				ResId: resid.FromString(fmt.Sprintf("PolicySet../%s.", policySetConfig.Name)),
			},
			Patch: fmt.Sprintf(`[{"op": "replace", "path": "/metadata/annotations/%s", "value": "%s"}]`, pkg.ANNOTATION_COMPONENT_TITLE, config.PolicySetName),
		}
		policySetPatches = append(policySetPatches, policySetPatch)
	}

	policyDefaults := pgtype.PolicyDefaults{
		Namespace: config.Namespace,
		PolicyOptions: pgtype.PolicyOptions{
			Placement: pgtype.PlacementConfig{
				LabelSelector: config.clusterSelectors,
			},
		},
		ConfigurationPolicyOptions: pgtype.ConfigurationPolicyOptions{
			NamespaceSelector: pgtype.NamespaceSelector{
				Exclude: []string{"kube-system", "open-cluster-management", "open-cluster-management-agent", "open-cluster-management-agent-addon"},
				Include: []string{"*"},
			},
		},
	}
	policyConfigs := []pgtype.PolicyConfig{}
	for _, policyConfig := range policyConfigMap {
		policyConfigs = append(policyConfigs, policyConfig)
	}
	policySetGeneratorManifest := policygenerator.BuildPolicyGeneratorManifest("policy-set", policyDefaults, policyConfigs)
	policySetGeneratorManifest.PlacementBindingDefaults.Name = "policy-set"
	policySetGeneratorManifest.PolicySets = policySets
	policySetGeneratorManifest.PolicySetDefaults = pgtype.PolicySetDefaults{
		PolicySetOptions: pgtype.PolicySetOptions{
			Placement: policyDefaults.Placement,
		},
	}

	if policySetGeneratorManifest.PolicyDefaults.Namespace == "" {
		policySetGeneratorManifest.PolicyDefaults.Namespace = DummyNamespace
	}
	if err := pkg.WriteObjToYamlFileByGoYaml(c.tempDir.GetTempDir()+"/policy-generator.yaml", policySetGeneratorManifest); err != nil {
		return err
	}

	logger.Info("Create configmap for templatized parameters")
	parametersConfigmap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "c2p-parameters",
			Namespace: config.Namespace,
		},
		Data: parameters,
	}
	if err := pkg.WriteObjToYamlFile(c.tempDir.GetTempDir()+"/parameters.yaml", parametersConfigmap); err != nil {
		return err
	}

	kustomize := typekustomize.Kustomization{
		Generators: []string{"./policy-generator.yaml"},
		Resources:  []string{"./parameters.yaml"},
		Patches:    policySetPatches,
	}
	if err := pkg.WriteObjToYamlFile(c.tempDir.GetTempDir()+"/kustomization.yaml", kustomize); err != nil {
		return err
	}
	return nil
}

func (c *Composer) CopyAllTo(destDir string) error {
	if _, err := pkg.MakeDir(destDir); err != nil {
		return err
	}
	if err := cp.Copy(c.tempDir.GetTempDir(), destDir); err != nil {
		return err
	}
	return nil
}

func (c *Composer) GeneratePolicySet() (*resmap.ResMap, error) {
	generatedManifests, err := policygenerator.Kustomize(c.tempDir.GetTempDir())
	if err != nil {
		logger.Error(err.Error(), "failed to run kustomize")
		return nil, err
	}
	// TODO: Workaround to allow to run PolicyGenerator with empty namespace.
	for _, resource := range generatedManifests.Resources() {
		if resource.GetNamespace() == DummyNamespace {
			if err := resource.SetNamespace(""); err != nil {
				return nil, err
			}
		}
	}
	return &generatedManifests, nil
}

func toDNSCompliant(name string) string {
	var result string
	result = strings.ToLower(name)
	result = strings.ReplaceAll(result, " ", "-")
	return result
}

func appendUnique(slice []string, elems ...string) []string {
	a := append(slice, elems...)
	return sets.List[string](sets.New[string](a...))
}

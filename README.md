# compliance-to-policy-go


# <img alt="Logo" width="50px" src="https://raw.githubusercontent.com/oscal-compass/compliance-to-policy/main/assets/compliance-to-policy-800x800.PNG" style="vertical-align: middle;" /> Compliance-to-Policy (also known as `C2P`)

Compliance-to-Policy (C2P) is designed to bridge Compliance as Code such as Open Security Controls Assessment Language (OSCAL) and Policy as Code used by Policy Validation Point (PVP).
It generates policies in native format of PVP from OSCAL Component Definitions and produces OSCAL Assessment Results from the native assessment results of PVP. C2P can be used both as a command-line tool and a Python library, making it easy and flexible to integrate into your Continuous Compliance pipelines, such as GitHub Actions, Tekton Pipelines, or Agile Authoring Pipelines. It supports multiple PVP engines, including [Kyverno](https://kyverno.io/) and [Open Cluster Management Policy Framework](https://open-cluster-management.io/).

![C2P Overview](/assets/architecture.png)

1. Compliance-to-Policy (C2P) is running in GitOps Pipeline, Kubernetes controller, or Python/Go environment
2. C2P receives Compliance as Code, for example OSCAL Component Definition that represents mapping between controls and policies (policy names/ids)
3. C2P generates policies through plugin for each policy engine
    - The plugin is responsible for implementing a function that takes policy names/ids and returns policies
4. Policies are delivered to policy engines by GitOps sync, the subsequence pipeline task, Kubernetes controller, or a deployment automation program 
5. Results are collected from policy engines by a scheduled task or Kubernetes controller
6. C2P aggregates the results of policy engines by controls through plugin for each policy engine
    - The plugin is responsible for implementing a function that takes the results of the policy engine and returns verdicts (pass/fail/error), reason, and/or resource name for each respective policy by its names/IDs.
7. C2P produces Compliance Assessment Results, for example OSCAL Assessment Results that represents the assessment results of each control

## Goals
Provide seamless integration with compliance frameworks and existing policy engines, and enable to use heterogeneous policy engines in compliance check operation
- Flexibility in choice of policy engines and compliance frameworks
    - Provide plugins to cover various policy engines including proprietary/open source policy validation/enforcement engines, or in-house policy validation/enforcement program
    - Cover various compliance frameworks not only OSCAL but also other GRC frameworks and Cloud Security Posture Management services
- Community-driven plugin extension
    - Provide an efficient plugin interface and development method

## Supported Compliance Frameworks
- [Open Security Controls Assessment Language (OSCAL)](https://pages.nist.gov/OSCAL/documentation/)
    - OSCAL standard provides a compliance framework and the corresponding set of key compliance artifacts expressed in machine processable formats enabling all compliance documents to be treated as code and therefore processed and managed in the same manner.

## Supported Policy Engines
- [Kyverno](https://kyverno.io/) (for Kubernetes resources)
    - Kyverno is a policy engine designed for Kubernetes, where policies are managed as Kubernetes resources. Kyverno policies can validate, mutate, generate, and clean up Kubernetes resources.
- [Open Cluster Management Policy Framework](https://open-cluster-management.io/) (for Kubernetes resources)
    - OCM is a multi-cluster management platform that provides governance of Kubernetes policies. [Its policy framework](https://open-cluster-management.io/concepts/policy/) allows for the validation and enforcement of policies across multiple clusters.

## Prerequisite
- Install [Kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/binaries/)
- Install [policy-generator plugin](https://github.com/open-cluster-management-io/policy-generator-plugin?tab=readme-ov-file)

## Usage of C2P CLI
```
$ c2pcli -h        
C2P CLI

Usage:
  c2pcli [flags]
  c2pcli [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  kyverno     C2P CLI Kyverno plugin
  ocm         C2P CLI OCM plugin
  version     Display version

Flags:
  -h, --help   help for c2pcli

Use "c2pcli [command] --help" for more information about a command.
```

C2P is targeting a plugin architecture to cover not only OCM Policy Framework but also other types of PVPs. 
Please go to the docs for each usage.
- [C2P for OCM](/docs/ocm/README.md) 
- [C2P for Kyverno](/docs/kyverno/README.md) 

## Build at local
```
make build
```
```
./bin/c2pcli_<version>_<os>_<arch> -h
```

## Test
```
make test
```

## Release
1. Create a git tag of the following format `go/<version>` (e.g. `go/v0.1.2`)
1. Run release command
    ```
    echo $PAT | gh auth login --with-token -h github.com
    make release 
    ```

## License & Authors

If you would like to see the detailed LICENSE click [here](LICENSE).
Consult [contributors](https://github.com/oscal-compass/compliance-to-policy/graphs/contributors) for a list of authors and [maintainers](MAINTAINERS.md) for the core team.

```text
# Copyright (c) 2024 The OSCAL Compass Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

```
______________________________________________________________________

We are a Cloud Native Computing Foundation sandbox project.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://www.cncf.io/wp-content/uploads/2022/07/cncf-white-logo.svg">
  <img src="https://www.cncf.io/wp-content/uploads/2022/07/cncf-color-bg.svg" width=300 />
</picture>

The Linux Foundation® (TLF) has registered trademarks and uses trademarks. For a list of TLF trademarks, see [Trademark Usage](https://www.linuxfoundation.org/legal/trademark-usage)".

*Compliance to Policy was originally created by IBM.*
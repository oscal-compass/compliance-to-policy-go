## C2P for OCM

### Usage of C2P CLI
```
C2P CLI

Usage:
  c2pcli [command]

Available Commands:
  completion    Generate the autocompletion script for the specified shell
  help          Help about any command
  oscal2policy  Transform OSCAL to policy artifacts.
  oscal2posture Generate Compliance Posture from OSCAL artifacts.
  result2oscal  Transform policy result artifacts to OSCAL Assessment Results.
  version       Display version

Flags:
      --debug   Run with debug log level
  -h, --help    help for c2pcli

Use "c2pcli [command] --help" for more information about a command.
```

### Prerequisites
1. Install [Policy Generator Plugin](https://github.com/open-cluster-management-io/policy-generator-plugin#as-a-kustomize-plugin)
1. Prepare OCM Policy Resources
    - You can use [policy-resources for test](/pkgstdata/ocm/policies)
    - You can also use [Policy Collection](https://github.com/open-cluster-management-io/policy-collection). Please see [C2P Decomposer](#c2p-decomposer)

### Manual end-to-end use case

#### Outline
1. Create OSCAL Component Definition
    - Use example one. In real cases, a user writes OSCAL by Authoring tool like [Trestle](https://github.com/oscal-compass/compliance-trestle))
2. Run oscal2policy to generate OCM Policies from OSCAL
3. Deploy generated OCM Policies to OCM Hub
4. Get OCM Policies from OCM Hub
5. Run result2oscal to generate OSCAL Assessment Results from the OCM Policy Results
6. Prettify OSCAL Assessment Results (See [tools](../tools))

![manual-end-to-end-use-case.png](/docs/ocm/images/manual-end-to-end-use-case.png)

#### Steps
1. Prerequisites
    1. OCM is configured to manage two k8s clusters (cluster1 and cluster2) and installed Policy Governance Framework.
    2. Namespace `c2p` is created in OCM Hub
    3. The managed clusters are labeled `my-cluster=true` and bound to `c2p` namespace
        ```
        $ clusteradm get clustersets

        <ManagedClusterSet> 
        └── <default> 
        │   ├── <BoundNamespace> 
        │   ├── <Status> 2 ManagedClusters selected
        │   ├── <Clusters> [cluster1 cluster2]
        └── <global> 
        │   ├── <BoundNamespace> 
        │   ├── <Status> 2 ManagedClusters selected
        │   ├── <Clusters> [cluster1 cluster2]
        └── <myclusterset> 
            └── <BoundNamespace> c2p
            └── <Status> 2 ManagedClusters selected
            └── <Clusters> [cluster1 cluster2]
        ```
2. Create the OCM manifest and place your plugin in the plugin directory
```bash
cp ../../bin/ocm-plugin ../../c2p-plugins
checksum=$(sha256sum ../../c2p-plugins/ocm-plugin | cut -d ' ' -f 1 )
cat > ../../c2p-plugins/c2p-ocm-manifest.json << EOF
{
 "metadata": {
   "id": "ocm",
   "description": "OCM PVP Plugin",
   "version": "0.0.1",
   "types": [
     "pvp"
   ]
 },
 "executablePath": "ocm-plugin",
 "sha256": "$checksum",
 "configuration": [
   {
     "name": "policy-dir",
     "description": "A directory where ocm policies are located.",
     "required": true
   },
   {
     "name": "policy-results-dir",
     "description": "A directory where policy results are located",
     "required": true
   },
   {
     "name": "temp-dir",
     "description": "A temporary directory for policies",
     "required": true
   },
   {
     "name": "output-dir",
     "description": "The output directory for policies",
     "required": false,
     "default": "."
   },
   {
      "name": "policy-set-name",
      "required": true
   },
   {
      "name": "namespace",
      "required": true
   }
 ]
}
EOF
```

1. Run oscal2policy to generate OCM Policies from OSCAL
    ```
    c2pcli oscal2policy -c ./docs/ocm/c2p-config.yaml -n nist_800_53
    ```
    - The generated ocm-policies directory looks like [./final-outputs/ocm-policies](final-outputs/ocm-policies)
2. Deploy the generated OCM Policies to OCM Hub
    ```
    kubectl create -f /tmp/ocm-policies
    ```
3. Wait for policies to be delivered
    ```
    $ kubectl get policy -A
    NAMESPACE   NAME                                        REMEDIATION ACTION   COMPLIANCE STATE   AGE
    c2p         policy-deployment                           inform               NonCompliant       5m15s
    c2p         policy-high-scan                            inform               NonCompliant       5m15s
    c2p         policy-install-kyverno-from-manifests       enforce              Compliant          5m14s
    c2p         policy-kyverno-require-labels                                    NonCompliant       5m14s
    cluster1    c2p.policy-deployment                       inform               NonCompliant       2m15s
    cluster1    c2p.policy-high-scan                        inform               NonCompliant       2m15s
    cluster1    c2p.policy-install-kyverno-from-manifests   enforce              Compliant          2m14s
    cluster1    c2p.policy-kyverno-require-labels                                NonCompliant       2m11s
    cluster2    c2p.policy-deployment                       inform               NonCompliant       2m15s
    cluster2    c2p.policy-high-scan                        inform               NonCompliant       2m15s
    cluster2    c2p.policy-install-kyverno-from-manifests   enforce              Compliant          2m14s
    cluster2    c2p.policy-kyverno-require-labels                                NonCompliant       2m11s
    ```
4. Get OCM Policy Results (Policy, PolicySet, PlacementDecision) from OCM Hub
    ```
    mkdir -p /tmp/results
    kubectl get policies.policy.open-cluster-management.io -A -o yaml > /tmp/results/policies.policy.open-cluster-management.io.yaml
    kubectl get policysets.policy.open-cluster-management.io -A -o yaml > /tmp/results/policysets.policy.open-cluster-management.io.yaml
    kubectl get placementdecisions.cluster.open-cluster-management.io -A -o yaml > /tmp/results/placementdecisions.cluster.open-cluster-management.io.yaml
    ```
5. Run result2oscal to generate OSCAL Assessment Results from the OCM Policy Results
    ```
    c2pcli result2oscal -c ./docs/ocm/c2p-config.yaml -n nist_800_53 -o /tmp/assessment-results.json
    ```
6. Prettify OSCAL Assessment Results in .md format
   ```bash
   c2pcli oscal2posture -c ./docs/ocm/c2p-config.yaml --assessment-results /tmp/assessment-results.json -o /tmp/compliance-posture.md
   ```
You can view an example compliance posture like [compliance-posture.md](../ocm/final-outputs/compliance-posture.md)

### GitOps automation use case

#### Outline

https://github.com/IBM/compliance-to-policy/assets/113283236/da3518d0-53de-4bd6-8703-04ce94e9dfba

#### Steps

Setup Github Repos

1. Create two repositories (one is configuration repository that's used for pipeline from OSCAL to Policy and another is evidence repository that's used for pipeline from OCM statuses to Compliance result)
    - For example, c2p-for-ocm-pipeline01-config and c2p-for-ocm-pipeline01-evidence
2. Create Github Personal Access Token having following permissions
    - Repository permission of `Contents`, `Pull Requests`, and `Workflows` with read-and-write against both the configuration repository and the evidence repository.
3. Fork C2P repository (yana1205/compliance-to-policy.git) and checkout `template`
4. Set required parameters for github action to initialize your configuration and evidence repo
    1. Go to Settings tab
    2. Go to `Actions` under `Secrets and variables`
    3. Create `New repository secret`
        - Name: PAT
        - Secret: Created Github Personal Access Token  
    4. Go to `Variables` tab to create `New repository variable`
    5. Create `CONFIGURATION_REPOSITORY` variable
        - Name: CONFIGURATION_REPOSITORY
        - Value: `<configuration repository org>/<configuration repository name> (e.g. yana1205/c2p-for-ocm-pipeline01-config)`
    6. Create `EVIDENCE_REPOSITORY` variable
        - Name: EVIDENCE_REPOSITORY
        - Value: `<evidence repository org>/<evidence repository name> (e.g. yana1205/c2p-for-ocm-pipeline01-evidence)`
5. Run Action `Initialize repositories` with branch `template`
6. Go to the configuration repository and create `New repository secret`
    - Name: PAT
    - Secret: Created Github Personal Access Token
7. Go to the evidence repository and create `New repository secret`
    - Name: PAT
    - Secret: Created Github Personal Access Token

Run oscal-to-pocliy

1. Go to the configuration repository
2. Go to `Actions` tab
3. Run `OSCAL to Policy`
    1. This action generates manifests from OSCAL and then generate a PR of changes for a directory `ocm-policy-manifests` containing the generated manifests.
4. Merge the PR

Integrate with GitOps

1. Sync `ocm-policy-manifests` directory with your OCM Hub by OCM GitOps (OCM Channel and Subscription addon)

Deploy collector to your OCM Hub

1. Apply RBAC for collector
    ```
    kubectl apply -f https://raw.githubusercontent.com/yana1205/compliance-to-policy/redesign.0622/scripts/collect/rbac.yaml
    ```
2. Create Secret for Github access
    ```
    kubectl -n c2p create secret generic --save-config collect-ocm-status-secret --from-literal=user=<github user> --from-literal=token=<github PAT> --from-literal=org=<evidence org name> --from-literal=repo=<evidence repo name>
    ```
    e.g.
    ```
    kubectl -n c2p create secret generic --save-config collect-ocm-status-secret --from-literal=user=yana1205 --from-literal=token=github_pat_xxx --from-literal=org=yana1205 --from-literal=repo=c2p-for-ocm-pipeline01-evidence
    ```
3. Deploy collector cronjob
    ```
    kubectl apply -f https://raw.githubusercontent.com/IBM/compliance-to-policy/main/scripts/collect/cronjob.yaml
    ```

Cleanup

```
kubectl delete -f https://raw.githubusercontent.com/IBM/compliance-to-policy/main/scripts/collect/cronjob.yaml
kubectl -n c2p delete secret collect-ocm-status-secret 
kubectl delete -f https://raw.githubusercontent.com/IBM/compliance-to-policy/main/scripts/collect/rbac.yaml
```


### C2P Decomposer
Decompose OCM poicy collection to kubernetes resources composing each OCM policy (we call it policy resource).

1. Clone [Policy Collection](https://github.com/open-cluster-management-io/policy-collection)
    ```
    git clone --depth 1 https://github.com/open-cluster-management-io/policy-collection.git /tmp/policy-collection
    ```
2. Run C2P Decomposer
    ```
    go run ./cmd/decompose/decompose.go --policy-collection-dir=/tmp/policy-collection --out=/tmp/c2p-output
    ```
3. Decomposed policy resources are ouput in `/tmp/c2p-output/decomposed/resources`
    ```
    $ tree -L 1 /tmp/c2p-output/decomposed
    /tmp/c2p-output/decomposed
    ├── _sources
    └── resources
    ```
    Individual decomposed resource contains k8s manifests and configuration files (policy-generator.yaml and kustomization.yaml) for PolicyGenerator. 
    ```
    $ tree -L 3 /tmp/c2p-output/decomposed/resources
    /tmp/c2p-output/decomposed/resources
    ├── add-chrony
    │   ├── add-chrony-worker
    │   │   └── MachineConfig.50-worker-chrony.0.yaml
    │   ├── kustomization.yaml
    │   └── policy-generator.yaml
    ├── add-tvk-license
    │   ├── add-tvk-license
    │   │   └── License.triliovault-license.0.yaml
    │   ├── kustomization.yaml
    ```
### C2P Composer
Compose OCM Policy from policy resources from compliance information (for example, [compliance.yaml](/cmd/compose/compliance.yaml))

1. Run C2P Composer
    ```
    go run cmd/compose-by-c2pcr/main.go --c2pcr ./cmd/compose-by-c2pcr/c2pcr.yaml --out /tmp/c2p-output
    ```
2. Composed OCM policies are output in `/tmp/c2p-output`
    ```
    $ tree /tmp/c2p-output                                                                             
    /tmp/c2p-output
    ├── add-chrony
    │   ├── add-chrony-worker
    │   │   └── MachineConfig.50-worker-chrony.0.yaml
    │   ├── kustomization.yaml
    │   └── policy-generator.yaml
    ├── install-odf-lvm-operator
    │   ├── kustomization.yaml
    │   ├── odf-lvmcluster
    │   │   └── LVMCluster.odf-lvmcluster.0.yaml
    │   ├── policy-generator.yaml
    │   └── policy-odf-lvm-operator
    │       ├── Namespace.openshift-storage.0.yaml
    │       ├── OperatorGroup.openshift-storage-operatorgroup.0.yaml
    │       └── Subscription.lvm-operator.0.yaml
    ├── kustomization.yaml
    ├── policy-generator.yaml
    └── policy-sets.yaml
    ```

## C2P for Kyverno

### Continuous Compliance by C2P 

https://github.com/oscal-compass/compliance-to-policy/assets/113283236/4b0b5357-4025-46c8-8d88-1f4c00538795

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

1. Prepare Kyverno Policy Resources
    - You can use [policy-resources for test](/pkgstdata/kyverno/policy-resources)
    - For bring your own policies, please see [Bring your own Kyverno Policy Resources](#bring-your-own-kyverno-policy-resources)

2. Create the Kyverno manifest and place your plugin in the plugin directory
```bash
cp ../../bin/kyverno-plugin ../../c2p-plugins
checksum=$(sha256sum ../../c2p-plugins/kyverno-plugin | cut -d ' ' -f 1 )
cat > ../../c2p-plugins/c2p-kyverno-manifest.json << EOF
{
  "metadata": {
    "id": "kyverno",
    "description": "Kyverno PVP Plugin",
    "version": "0.0.1",
    "types": [
      "pvp"
    ]
  },
  "executablePath": "kyverno-plugin",
  "sha256": "$checksum",
  "configuration": [
    {
      "name": "policy-dir",
      "description": "A directory where kyverno policies are located.",
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
    }
  ]
}
EOF
```


#### Convert OSCAL to Kyverno Policy
```
$ c2pcli oscal2policy -c docs/kyverno/c2p-config.yaml -n nist_800_53
2023-10-31T07:23:56.291+0900    INFO    kyverno/c2pcr   kyverno/configparser.go:53      Component-definition is loaded from ./pkg/testdata/kyverno/component-definition.json

$ tree /tmp/kyverno-policies 
/tmp/kyverno-policies
└── allowed-base-images
    ├── 02-setup-cm.yaml
    └── allowed-base-images.yaml
```

#### Convert Policy Report to OSCAL Assessment Results
```
$ c2pcli result2oscal -c docs/kyverno/c2p-config.yaml -n nist_800_53 -o /tmp/assessment-results

$ tree /tmp/assessment-results 
/tmp/assessment-results
└── assessment-results.json
```

# Copyright 2025 The OSCAL Compass Authors
# SPDX-License-Identifier: Apache-2.0


#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cp -f "${SCRIPT_DIR}/../bin/kyverno-plugin" "${SCRIPT_DIR}/../bin/ocm-plugin" "${SCRIPT_DIR}/../c2p-plugins"

checksum=$(sha256sum "${SCRIPT_DIR}/../c2p-plugins/kyverno-plugin" | cut -d ' ' -f 1 )
cat > "${SCRIPT_DIR}/../c2p-plugins/c2p-kyverno-manifest.json" << EOF
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


checksum=$(sha256sum "${SCRIPT_DIR}/../c2p-plugins/ocm-plugin" | cut -d ' ' -f 1 )
cat > "${SCRIPT_DIR}/../c2p-plugins/c2p-ocm-manifest.json" << EOF
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
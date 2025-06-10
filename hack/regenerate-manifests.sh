#!/bin/bash
# Copyright 2025 The OSCAL Compass Authors
# SPDX-License-Identifier: Apache-2.0

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PLUGINS_DIR="${SCRIPT_DIR}/../c2p-plugins"

if [ ! -d "${PLUGINS_DIR}" ]; then
  echo "Creating plugins directory: ${PLUGINS_DIR}"
  mkdir -p "${PLUGINS_DIR}"
fi

cp -f "${SCRIPT_DIR}/../bin/kyverno-plugin" "${SCRIPT_DIR}/../bin/ocm-plugin" "${PLUGINS_DIR}"

checksum=$(sha256sum "${PLUGINS_DIR}/kyverno-plugin" | cut -d ' ' -f 1 )
cat > "${PLUGINS_DIR}/c2p-kyverno-manifest.json" << EOF
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


checksum=$(sha256sum "${PLUGINS_DIR}/ocm-plugin" | cut -d ' ' -f 1 )
cat > "${PLUGINS_DIR}/c2p-ocm-manifest.json" << EOF
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
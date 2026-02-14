# GitHub Actions for C2P Go

## Setup C2PCLI

#### Usage

See [action.yml](./setup-c2pcli/action.yml)

```yaml
steps:
  - uses: actions/checkout@v4
  - uses: oscal-compass/compliance-to-policy-go/actions/setup-c2pcli@HEAD
    with:
      c2p-version: 'v2.0.0-alpha.X' # The C2P version to download
  - run: c2pcli version
```

> Note: Only GitHub provided `ubuntu` and `macos` based runner are currently supported. For self-hosted runners, you must in the GitHub CLI first.
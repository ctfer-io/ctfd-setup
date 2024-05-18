<div align="center">
  <h1>CTFd-Setup</h1>
  <p><b>Version your CTFd setup configuration.</b><p>
  <a href="https://pkg.go.dev/github.com/ctfer-io/ctfd-setup"><img src="https://shields.io/badge/-reference-blue?logo=go&style=for-the-badge" alt="reference"></a>
  <a href=""><img src="https://img.shields.io/github/license/ctfer-io/ctfd-setup?style=for-the-badge" alt="License"></a>
  <a href="https://coveralls.io/github/ctfer-io/ctfd-setup?branch=main"><img src="https://img.shields.io/coverallsCoverage/github/ctfer-io/ctfd-setup?style=for-the-badge" alt="Coverage Status"></a>
	<br>
	<a href="https://github.com/ctfer-io/ctfd-setup/actions/workflows/codeql-analysis.yaml"><img src="https://img.shields.io/github/actions/workflow/status/ctfer-io/ctfd-setup/codeql-analysis.yaml?style=for-the-badge&label=CodeQL" alt="CodeQL"></a>
    <a href="https://securityscorecards.dev/viewer/?uri=github.com/ctfer-io/ctfd-setup"><img src="https://img.shields.io/ossf-scorecard/github.com/ctfer-io/ctfd-setup?label=openssf%20scorecard&style=for-the-badge" alt="OpenSSF Scoreboard"></a>
</div>

CTFd does not have the concept of **configuration file**, leading to **deployment complications** and the **impossibility to version configurations**.
This is problematic for reproducibility or sharing configuration for debugging or replicating a CTF infrastructure.

Moreover, the setup API does not exist, so we had to map it to what the frontend calls in [go-ctfd](https://github.com/ctfer-io/go-ctfd/blob/main/api/setup.go).

To fit those gaps, we built `ctfd-setup` on top of the CTFd API. This utility helps setup a CTFd instance from a YAML configuration file, CLI flags and environment variables.
Thanks to this, you can integrate it using **GitHub Actions**, **Drone CI** or even as part of your **IaC provisionning**.

<!--TODO improve CI + CD support description-->

## How to use

<div align="center">
    <img src="res/how-to-use.excalidraw.png" alt="ctfd-setup utility used in GitHub Actions, Drone CI and Docker and Kubernetes initial container" width="800px">
</div>

For the CLI configuration, please refer to the binary's specific API through `ctfd-setup --help`.
In use of IaC provisionning scenario, the corresponding environment variables are also mapped to the output, so please refer to it.

### GitHub Actions

To improve our own workflows and share knownledges and tooling, we built a GitHub Action: `ctfer-io/ctfd-setup`.
You can use it given the following example.

```yaml
name: 'My workflow'

on:
  push:
    branches:
      - 'main'

jobs:
  my-job:
    runs-on: 'ubuntu-latest'
    steps:
      - name: 'Setup CTFd'
        uses: 'ctfer-io/ctfd-setup@v0'
        with:
          url: ${{ secrets.CTFD_URL }}
          appearance_name: 'My CTF'
          appearance_description: 'My CTF description'
          admin_name: ${{ secrets.ADMIN_USERNAME }}
          admin_email: ${{ secrets.ADMIN_EMAIL }}
          admin_password: ${{ secrets.ADMIN_PASSWORD }}
          # ... and so on (non-mandatory attributes)
```

### Drone CI

This could also be used as part of a Drone CI use `ctferio/ctfd-setup`.

```yaml
kind: pipeline
type: docker
name: 'My pipeline'

trigger:
  branch:
  - main
  event:
  - push

steps:
  # ...

  - name: 'Setup CTFd'
    image: 'ctferio/ctfd-setup@v0'
    settings:
      url:
        from_secret: CTFD_URL
      appearance_name: 'My CTF'
      appearance_description: 'My CTF description'
      admin_name:
        from_secret: ADMIN_USERNAME
      admin_email:
        from_secret: ADMIN_EMAIL
      admin_password:
        from_secret: ADMIN_PASSWORD
      # ... and so on (non-mandatory attributes)
```

# Reporting Security Issues

Please report any security issues you discovered in the CTFd setup binary to ctfer-io@protonmail.com.

We will assess the risk, plus make a fix available before we create a GitHub issue.

In case the vulnerability is into a dependency, please refer to their respective security policy directly.

Thank you for your contribution.

## Refering to this repository

To refer to this repository using a CPE v2.3, please use `cpe:2.3:a:ctfer-io:ctfd-setup:*:*:*:*:*:*:*:*` with the `version` set to the tag you are using.

### Signature and Attestations

For deployment purposes (and especially in the deployment case of Kubernetes), you may want to ensure the integrity of what you run.

The release assets are SLSA 3 and can be verified using [slsa-verifier](https://github.com/slsa-framework/slsa-verifier) using the following.

```bash
slsa-verifier verify-artifact "<path/to/release_artifact>"  \
  --provenance-path "<path/to/release_intoto_attestation>"  \
  --source-uri "github.com/ctfer-io/ctfd-setup" \
  --source-tag "<tag>"
```

The Docker image is SLSA 3 and can be verified using [slsa-verifier](https://github.com/slsa-framework/slsa-verifier) using the following.

```bash
slsa-verifier slsa-verifier verify-image "ctferio/ctfd-setup:<tag>@sha256:<digest>" \
    --source-uri "github.com/ctfer-io/ctfd-setup" \
    --source-tag "<tag>"
```

Alternatives exist, like [Kyverno](https://kyverno.io/) for a Kubernetes-based deployment.

### SBOMs

A SBOM for the whole repository is generated on each release and can be found in the assets of it.
They are signed as SLSA 3 assets. Refer to [Signature and Attestations](#signature-and-attestations) to verify their integrity.

A SBOM is generated for the Docker image in its manifest, and can be inspected using the following.

```bash
docker buildx imagetools inspect "ctferio/ctfd-setup:<tag>" \
    --format "{{ json .SBOM.SPDX }}"
```

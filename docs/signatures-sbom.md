# Signatures and SBOM

This document describes the signature and Software Bill of Materials (SBOM) functionality implemented in the git-sync service supply chain.

## Supply Chain Security Overview

Our supply chain security implementation follows these steps:

1. **Build**: Multi-architecture Docker images using Docker Buildx
2. **Push**: Images pushed to container registries (GHCR and Docker Hub)
3. **Sign**: Images signed with cosign (keyless signing)
4. **SBOM**: Generate SBOM with syft in CycloneDX format
5. **Attach**: Attach SBOM to image as OCI artifact
6. **Attest**: Create SLSA provenance attestation
7. **Scan**: Vulnerability scanning with Trivy

## Image Signing

We use cosign for keyless signing of Docker images. This approach doesn't require managing private keys, instead leveraging identity tokens from the CI environment.

### Verifying Image Signatures

To verify the signature of a published image:

```bash
# Install cosign first
# Then verify the image signature
cosign verify ghcr.io/your-username/git-sync:tag \
  --certificate-identity-regexp https://github.com/your-username/git-sync/.github/workflows/ \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

## Software Bill of Materials (SBOM)

We generate SBOMs in CycloneDX format using syft. These SBOMs provide a detailed inventory of all software components included in the Docker image.

### Generating SBOM

The SBOM is automatically generated during the CI/CD process, but you can also generate it locally:

```bash
# Install syft first
# Generate SBOM for a built image
syft packages ghcr.io/your-username/git-sync:tag -o cyclonedx-json > sbom.cdx.json
```

### Verifying SBOM

To verify the SBOM attached to an image:

```bash
# Download the attached SBOM
cosign download sbom ghcr.io/your-username/git-sync:tag

# Verify the SBOM signature
SBOM_REFERENCE=$(cosign triangulate --type sbom ghcr.io/your-username/git-sync:tag)
cosign verify $SBOM_REFERENCE \
  --certificate-identity-regexp https://github.com/your-username/git-sync/.github/workflows/ \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

## SLSA Provenance

We generate SLSA (Supply-chain Levels for Software Artifacts) provenance attestations using GitHub's official action. This provides verifiable build metadata.

### Verifying Provenance

To verify the SLSA provenance attestation:

```bash
# Verify the provenance attestation
gh attestation verify oci://ghcr.io/your-username/git-sync:tag \
  --owner your-username
```

## Vulnerability Scanning

We use Trivy to scan our images for vulnerabilities during the CI/CD process. Images are scanned for critical and high severity vulnerabilities, which will fail the build if found.

### Local Scanning

You can scan images locally with Trivy:

```bash
# Install trivy first
# Scan an image
trivy image ghcr.io/your-username/git-sync:tag
```

## Version Synchronization

The `/version` HTTP endpoint and the `git_sync_build_info` Prometheus metric are synchronized to provide consistent version information:

- Both use the same version information injected at build time
- Both include the same metadata: version, commit, date, and dirty status
- This ensures consistency between API responses and monitoring metrics

## Key Decisions

### CycloneDX for SBOM

We chose CycloneDX format for our SBOMs because:

1. It's a widely adopted standard for software supply chain security
2. It provides detailed component information and dependency relationships
3. It's supported by many security tools and platforms
4. It has good integration with syft and other CNCF projects

### Keyless Cosign

We use keyless cosign because:

1. It eliminates the need to manage private keys
2. It leverages the CI/CD environment's identity for signing
3. It's more secure as private keys never leave the signing environment
4. It's easier to maintain and doesn't require key rotation

## Commands Summary

| Command | Description |
|---------|-------------|
| `cosign verify IMAGE` | Verify image signature |
| `cosign download sbom IMAGE` | Download attached SBOM |
| `syft packages IMAGE` | Generate SBOM locally |
| `gh attestation verify IMAGE` | Verify SLSA provenance |
| `trivy image IMAGE` | Scan for vulnerabilities |

## Further Reading

- [Cosign Documentation](https://docs.sigstore.dev/cosign/overview/)
- [Syft Documentation](https://github.com/anchore/syft)
- [SLSA Specifications](https://slsa.dev/spec/v1.0/about)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [CycloneDX Specification](https://cyclonedx.org/specification/overview/)
# 1. CycloneDX for SBOM and Keyless Cosign

Date: 2025-09-27

## Status

Accepted

## Context

We need to implement supply chain security for the git-sync service, including:
1. Software Bill of Materials (SBOM) generation
2. Image signing for verification of artifact integrity

Several options exist for both SBOM formats and signing mechanisms, each with different trade-offs in terms of industry adoption, tooling support, and security properties.

## Decision

We will use:
1. **CycloneDX** as the SBOM format
2. **Keyless Cosign** for image signing

## Consequences

### Positive

#### CycloneDX for SBOM
- Industry standard with wide adoption in security tools
- Comprehensive component and dependency information
- Good integration with syft and other CNCF projects
- Supported by major security platforms and scanners
- Detailed vulnerability correlation capabilities

#### Keyless Cosign
- Eliminates private key management complexity
- Leverages CI/CD environment identity for signing
- More secure as private keys never leave the signing environment
- Easier maintenance without key rotation requirements
- OIDC-based identity verification provides strong security guarantees

### Negative

#### CycloneDX for SBOM
- Larger file size compared to SPDX in some cases
- May include more information than needed for simple use cases
- Tooling ecosystem, while growing, is not as mature as SPDX

#### Keyless Cosign
- Requires specific CI/CD environment support (GitHub Actions, GitLab CI, etc.)
- Dependent on OIDC provider availability and reliability
- Short-lived certificates require online verification
- May not be suitable for air-gapped environments

## Considered Options

### SBOM Formats

1. **CycloneDX**
   - Pros: Wide industry adoption, comprehensive data model, good tooling support
   - Cons: Larger file sizes, newer standard than SPDX

2. **SPDX**
   - Pros: ISO standard, mature ecosystem, smaller file sizes
   - Cons: Less detailed dependency information, less adoption in security tools

3. **Syft Native Format**
   - Pros: Optimized for syft, fastest generation
   - Cons: Limited tooling support, not an industry standard

### Signing Mechanisms

1. **Keyless Cosign**
   - Pros: No key management, identity-based signing, automatic certificate renewal
   - Cons: Requires OIDC provider, online verification needed

2. **Traditional Cosign with Keys**
   - Pros: Works in any environment, offline verification possible
   - Cons: Key management complexity, rotation requirements, security risks

3. **Notary Project**
   - Pros: CNCF project, strong specification
   - Cons: More complex setup, less tooling maturity

## Decision Drivers

1. **Security**: Strong security properties with minimal operational overhead
2. **Adoption**: Industry standard tools and formats
3. **Maintainability**: Minimal ongoing operational burden
4. **Integration**: Good tooling support in our CI/CD environment
5. **Future-proofing**: Alignment with emerging supply chain security standards

## Related ADRs

- ADR-002: Multi-architecture Docker Images
- ADR-003: Container Base Image Selection

## Notes

This decision aligns with the broader supply chain security recommendations from the SLSA framework and follows industry best practices for software artifact verification.
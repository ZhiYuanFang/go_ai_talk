# runtime-docs-centralization-and-governance Specification

## Purpose
TBD - created by archiving change worker-dedicated-config-and-dao-simplification. Update Purpose after archive.
## Requirements
### Requirement: Runtime docs SHALL be centralized and governed
`dao-sync-by-domain.md` and `release-deploy-and-run.md` MUST be maintained in a dedicated runtime-docs directory, and change governance MUST require synchronized updates when runtime behavior changes.

#### Scenario: Docs location is consolidated
- **WHEN** checking runtime operation documents
- **THEN** both target documents MUST exist under one dedicated new folder

#### Scenario: Governance requires synchronized update
- **WHEN** project runtime/deployment/database boundary rules change
- **THEN** project governance (`openspec/project.md`) MUST require updating both runtime docs


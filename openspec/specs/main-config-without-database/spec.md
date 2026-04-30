# main-config-without-database Specification

## Purpose
TBD - created by archiving change worker-dedicated-config-and-dao-simplification. Update Purpose after archive.
## Requirements
### Requirement: Main config SHALL not carry database settings
`manifest/config/config.yaml` MUST NOT contain `database.*` fields after worker dedicated configuration is introduced.

#### Scenario: Review main config fields
- **WHEN** auditing `manifest/config/config.yaml`
- **THEN** no database connection configuration MUST exist in the file

#### Scenario: Gateway runtime without DB dependency
- **WHEN** `gateway-service` starts with main config
- **THEN** gateway MUST run without requiring database fields from main config


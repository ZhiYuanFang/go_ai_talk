# dao-extension-layer-simplification Specification

## Purpose
TBD - created by archiving change worker-dedicated-config-and-dao-simplification. Update Purpose after archive.
## Requirements
### Requirement: DAO extension files SHALL follow minimum-necessary rule
DAO `*_ext.go` files MUST be retained only when they provide business-meaningful extensions beyond generated DAO wrappers.

#### Scenario: Ext file has no added behavior
- **WHEN** an ext file only duplicates generated DAO behavior without business logic
- **THEN** the file MUST be merged away or removed

#### Scenario: Ext file provides service-specific query semantics
- **WHEN** an ext file includes domain query composition or behavior not present in generated DAO
- **THEN** the file MAY be retained with explicit comment/documented rationale


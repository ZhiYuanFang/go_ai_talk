# worker-dedicated-config-loading Specification

## Purpose
TBD - created by archiving change worker-dedicated-config-and-dao-simplification. Update Purpose after archive.
## Requirements
### Requirement: Worker-service SHALL use dedicated configuration
`worker-service` MUST have a dedicated default config file and MUST load it when `GF_GCFG_FILE` is not explicitly provided.

#### Scenario: Worker starts without GF_GCFG_FILE
- **WHEN** `worker-service` starts and `GF_GCFG_FILE` is empty
- **THEN** runtime MUST default to `manifest/config/config.worker-service.yaml`

#### Scenario: Deployment manifest uses worker dedicated config
- **WHEN** compose/kustomize/dockerfile defines worker runtime env
- **THEN** worker `GF_GCFG_FILE` MUST point to `manifest/config/config.worker-service.yaml`


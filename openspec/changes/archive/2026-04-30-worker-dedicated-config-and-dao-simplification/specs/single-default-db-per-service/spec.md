## ADDED Requirements

### Requirement: Service DAO access SHALL use only default database
Service processes MUST access database via their own `database.default` connection and MUST NOT rely on multi-group routing fallback logic.

#### Scenario: Domain DB group resolver removed
- **WHEN** checking DAO infrastructure files
- **THEN** `internal/dao/domain_db.go` multi-group resolver MUST be removed

#### Scenario: Service reads only local default connection
- **WHEN** a service executes DAO operations
- **THEN** the resolved DB connection MUST come from the service-local `database.default` config

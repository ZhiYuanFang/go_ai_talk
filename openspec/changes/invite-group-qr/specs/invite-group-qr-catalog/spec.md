## ADDED Requirements

### Requirement: Catalog returns inviteGroupQrUrl only while valid

When synthesizing `GET /cash/app/api/feature/catalog`, the system MUST include top-level `inviteGroupQrUrl` only when `expires_at > 0` and `expires_at > now`. The URL MUST point at the apk download path for `er_code.png` and SHOULD include a cache-busting query derived from `updated_at`. When invalid, the field MUST be omitted or empty.

#### Scenario: Valid window

- **WHEN** expiry is in the future and metadata exists
- **THEN** catalog data contains non-empty `inviteGroupQrUrl` referencing `er_code.png`

#### Scenario: Expired or zero

- **WHEN** `expires_at` is 0 or not after now
- **THEN** catalog MUST NOT present a usable invite group QR URL to the App

## ADDED Requirements

### Requirement: Admin can upload er_code.png to apk storage

The system MUST accept an Admin multipart upload that saves the image as exactly `er_code.png` under the gateway APK storage directory, overwriting any previous file. The system MUST reject uploads that are not a PNG when the implementation chooses reject-non-png. The download MUST remain available at `/device/app/apk/er_code.png` using the existing apk download route.

#### Scenario: Overwrite

- **WHEN** Admin uploads a new PNG as the group QR
- **THEN** the storage path contains only `er_code.png` as the group QR asset (prior content replaced)

### Requirement: Admin configures expiry and can preview after expiry

The system MUST allow Admin to set `expires_at` for the group QR in cash. When `expires_at` is 0 or in the past, App catalog MUST NOT expose the URL, but Admin preview MUST still be able to load the current `er_code.png`.

#### Scenario: Expired still previewable

- **WHEN** `expires_at` is in the past
- **THEN** Admin UI can still preview the image and see the expiry, and App catalog omits `inviteGroupQrUrl`

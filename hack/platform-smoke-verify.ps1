param(
    [string]$GatewayBase = "http://127.0.0.1:9701",
    [string]$HistoryBase = "http://127.0.0.1:9801",
    [string]$VoiceBase = "http://127.0.0.1:9802",
    [string]$DeviceBase = "http://127.0.0.1:9803",
    [string]$WorkerBase = "http://127.0.0.1:9901"
)

$ErrorActionPreference = "Stop"

function Assert-HttpOk([string]$url) {
    Write-Host "Checking $url"
    $resp = Invoke-WebRequest -Uri $url -Method Get -TimeoutSec 5
    if ($resp.StatusCode -lt 200 -or $resp.StatusCode -ge 300) {
        throw "Health check failed for $url, status=$($resp.StatusCode)"
    }
}

Assert-HttpOk "$GatewayBase/api.json"
Assert-HttpOk "$HistoryBase/api.json"
Assert-HttpOk "$VoiceBase/api.json"
Assert-HttpOk "$DeviceBase/api.json"
Assert-HttpOk "$WorkerBase/healthz"

Write-Host "Smoke verify passed."


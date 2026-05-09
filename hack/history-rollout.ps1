param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("local", "canary10", "canary50", "canary100", "proxy", "rollback")]
    [string]$Stage,

    [string]$HistoryProxyUrl = "http://127.0.0.1:9801",
    [switch]$PrintOnly
)

function Set-RolloutEnv {
    param(
        [string]$Mode,
        [string]$Percent
    )
    $env:HISTORY_API_PROXY_URL = $HistoryProxyUrl
    $env:HISTORY_API_ROUTE_MODE = $Mode
    $env:HISTORY_API_PROXY_CANARY_PERCENT = $Percent
}

switch ($Stage) {
    "local" {
        Set-RolloutEnv -Mode "local" -Percent "0"
    }
    "canary10" {
        Set-RolloutEnv -Mode "canary" -Percent "10"
    }
    "canary50" {
        Set-RolloutEnv -Mode "canary" -Percent "50"
    }
    "canary100" {
        Set-RolloutEnv -Mode "canary" -Percent "100"
    }
    "proxy" {
        Set-RolloutEnv -Mode "proxy" -Percent "100"
    }
    "rollback" {
        Set-RolloutEnv -Mode "local" -Percent "0"
    }
}

Write-Host "=== history rollout stage applied ==="
Write-Host ("Stage: " + $Stage)
Write-Host ("HISTORY_API_PROXY_URL=" + $env:HISTORY_API_PROXY_URL)
Write-Host ("HISTORY_API_ROUTE_MODE=" + $env:HISTORY_API_ROUTE_MODE)
Write-Host ("HISTORY_API_PROXY_CANARY_PERCENT=" + $env:HISTORY_API_PROXY_CANARY_PERCENT)

Write-Host ""
Write-Host "Next steps:"
Write-Host "1) restart gateway service to apply env vars."
Write-Host "2) verify endpoint:"
Write-Host "   GET /device/history/api/list?deviceNo=<sample>&page=1&pageSize=20"
Write-Host "3) observe error rate/latency logs before next stage."
Write-Host ""
Write-Host "Quick rollback command:"
Write-Host "   .\hack\history-rollout.ps1 -Stage rollback"

if ($PrintOnly) {
    Write-Host ""
    Write-Host "PrintOnly mode enabled. No extra actions executed."
}

param(
    [ValidateSet("local", "canary", "full", "rollback")]
    [string]$Stage = "local",
    [ValidateRange(0, 100)]
    [int]$CanaryPercent = 10
)

if ($Stage -eq "local") {
    $env:HISTORY_API_ROUTE_MODE = "local"
    $env:VOICE_API_ROUTE_MODE = "local"
    $env:DEVICE_API_ROUTE_MODE = "local"
}
elseif ($Stage -eq "canary") {
    if ($CanaryPercent -gt 0) {
        $env:HISTORY_API_ROUTE_MODE = "canary"
        $env:HISTORY_API_PROXY_CANARY_PERCENT = "$CanaryPercent"
        $env:VOICE_API_ROUTE_MODE = "proxy"
        $env:DEVICE_API_ROUTE_MODE = "proxy"
    }
}
elseif ($Stage -eq "full") {
    $env:HISTORY_API_ROUTE_MODE = "proxy"
    $env:VOICE_API_ROUTE_MODE = "proxy"
    $env:DEVICE_API_ROUTE_MODE = "proxy"
}
elseif ($Stage -eq "rollback") {
    $env:HISTORY_API_ROUTE_MODE = "local"
    $env:VOICE_API_ROUTE_MODE = "local"
    $env:DEVICE_API_ROUTE_MODE = "local"
}

Write-Host ("HISTORY_API_ROUTE_MODE=" + $env:HISTORY_API_ROUTE_MODE)
Write-Host ("VOICE_API_ROUTE_MODE=" + $env:VOICE_API_ROUTE_MODE)
Write-Host ("DEVICE_API_ROUTE_MODE=" + $env:DEVICE_API_ROUTE_MODE)


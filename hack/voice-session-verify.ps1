param(
    [string]$GatewayA = "http://127.0.0.1:9701",
    [string]$GatewayB = "http://127.0.0.1:9702",
    [string]$AdminPassword = "a521521521",
    [string]$DeviceNo = "device-verify-001"
)

$ErrorActionPreference = "Stop"

function Invoke-TextChat {
    param(
        [string]$BaseUrl,
        [string]$Transcript
    )
    $uri = ($BaseUrl.TrimEnd("/") + "/voice/text/chat")
    $headers = @{
        "X-Admin-Password" = $AdminPassword
        "Content-Type"     = "application/json"
    }
    $body = @{
        deviceNo   = $DeviceNo
        transcript = $Transcript
    } | ConvertTo-Json

    return Invoke-RestMethod -Method Post -Uri $uri -Headers $headers -Body $body
}

Write-Host "=== Voice session multi-instance verification ==="
Write-Host ("Gateway A: " + $GatewayA)
Write-Host ("Gateway B: " + $GatewayB)
Write-Host ("DeviceNo : " + $DeviceNo)
Write-Host ""

$r1 = Invoke-TextChat -BaseUrl $GatewayA -Transcript "请记住我是第一轮提问"
Write-Host "[A] reply:" $r1.reply

$r2 = Invoke-TextChat -BaseUrl $GatewayB -Transcript "这是第二轮提问，请延续上下文"
Write-Host "[B] reply:" $r2.reply

if ([string]::IsNullOrWhiteSpace($r1.reply) -or [string]::IsNullOrWhiteSpace($r2.reply)) {
    Write-Error "verification failed: empty reply detected"
    exit 1
}

Write-Host ""
Write-Host "Verification passed: two-instance text chat requests returned replies."
Write-Host "Next: run Redis node stop/start drill and repeat this script."

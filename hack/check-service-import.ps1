# 检查 internal/services 业务包之间是否互引（service-package-isolation）。
# 允许：同包、platform、contracts、aimodel、clients、标准库与第三方。
# 用法：powershell -File hack/check-service-import.ps1
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$domains = @("cash", "voice", "device", "history", "ucg", "gatewayapp", "simuser", "mcpbridge", "appstatus")
$fail = 0

function Get-ImportHits([string]$srcDir, [string]$dst) {
    if (-not (Test-Path $srcDir)) { return @() }
    $pattern = [regex]::Escape('"hello/internal/services/' + $dst + '"')
    $hits = @()
    Get-ChildItem -Path $srcDir -Recurse -Filter *.go | ForEach-Object {
        $lines = Select-String -Path $_.FullName -Pattern $pattern -SimpleMatch:$false -ErrorAction SilentlyContinue
        foreach ($l in $lines) {
            $hits += ("{0}:{1}:{2}" -f $_.FullName.Replace($root + "\", "").Replace($root + "/", ""), $l.LineNumber, $l.Line.Trim())
        }
    }
    return $hits
}

foreach ($src in $domains) {
    $srcDir = Join-Path "internal/services" $src
    if (-not (Test-Path $srcDir)) { continue }
    foreach ($dst in $domains) {
        if ($src -eq $dst) { continue }
        # gatewayapp 子包（如 usagestats）允许 import 同域 gatewayapp
        if ($src -eq "gatewayapp" -and $dst -eq "gatewayapp") { continue }
        $hits = Get-ImportHits $srcDir $dst
        if ($hits.Count -gt 0) {
            Write-Host "FAIL: services/$src 禁止 import services/$dst" -ForegroundColor Red
            $hits | ForEach-Object { Write-Host $_ -ForegroundColor Red }
            $fail = 1
        }
    }
}

if ($fail -eq 0) {
    Write-Host "OK: services 业务包互引检查通过"
}
exit $fail

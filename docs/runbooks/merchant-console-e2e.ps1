#Requires -Version 5.1
<#
.SYNOPSIS
  商户台 / 多宝宝端到端冒烟（任务 9.2 / 9.3）。

.DESCRIPTION
  对已部署且含 merchant-console-multi-baby 的网关执行：
  注册 → 免费 2 宝宝 → 第 3 拒绝 → staff 分发 → 欠费禁 create → Hub 授期 →
  第 3 宝宝 → create/end 历史 → subscribe + simulate-paid 延长 expire_at。

.PARAMETER BaseUrl
  网关根，如 http://127.0.0.1:8000 或 https://test.pangbao.cuplay.top

.PARAMETER AdminUser / AdminPassword
  Hub Admin 登录（grant 用）
#> 
param(
  [Parameter(Mandatory = $true)][string]$BaseUrl,
  [string]$AdminUser = "admin",
  [Parameter(Mandatory = $true)][string]$AdminPassword
)

$ErrorActionPreference = "Stop"
$BaseUrl = $BaseUrl.TrimEnd("/")
$suffix = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$store = "e2e_shop_$suffix"
$pass = "E2ePass!$suffix"
$staffAccount = "e2e_staff_$suffix"
$staffPass = "StaffPass!$suffix"

function Invoke-Api {
  param(
    [string]$Method,
    [string]$Path,
    [hashtable]$Headers = @{},
    [object]$Body = $null,
    [int[]]$AllowCodes = @()
  )
  $uri = "$BaseUrl$Path"
  $hdr = @{ "Content-Type" = "application/json" }
  foreach ($k in $Headers.Keys) { $hdr[$k] = $Headers[$k] }
  $params = @{
    Uri             = $uri
    Method          = $Method
    Headers         = $hdr
    UseBasicParsing = $true
    TimeoutSec      = 30
  }
  if ($null -ne $Body) {
    $params.Body = ($Body | ConvertTo-Json -Compress -Depth 6)
  }
  try {
    $resp = Invoke-WebRequest @params
    $code = [int]$resp.StatusCode
    $raw = $resp.Content
  } catch {
    $ex = $_.Exception
    if ($ex.Response) {
      $code = [int]$ex.Response.StatusCode
      $reader = New-Object System.IO.StreamReader($ex.Response.GetResponseStream())
      $raw = $reader.ReadToEnd()
      $reader.Close()
    } else {
      throw
    }
  }
  $json = $null
  if ($raw) {
    try { $json = $raw | ConvertFrom-Json } catch { $json = $null }
  }
  if ($AllowCodes -contains $code) {
    return @{ Status = $code; Json = $json; Raw = $raw }
  }
  if ($null -eq $json) {
    throw "HTTP $code empty/non-json: $Path :: $raw"
  }
  if ($json.code -ne 0) {
    throw "API code=$($json.code) msg=$($json.message) path=$Path"
  }
  return @{ Status = $code; Json = $json; Data = $json.data; Raw = $raw }
}

function Assert-True([bool]$Cond, [string]$Msg) {
  if (-not $Cond) { throw "ASSERT: $Msg" }
  Write-Host "OK  $Msg"
}

Write-Host "=== merchant-console E2E against $BaseUrl ==="

# 0) SPA 匿名壳
try {
  $spa = Invoke-WebRequest -Uri "$BaseUrl/merchant" -UseBasicParsing -TimeoutSec 20
  Assert-True ($spa.StatusCode -eq 200) "SPA /merchant HTTP 200"
} catch {
  throw "SPA 匿名加载失败（是否未部署本变更？）: $($_.Exception.Message)"
}

# 1) 商户注册
$reg = Invoke-Api -Method POST -Path "/device/app/api/merchant_register" -Body @{
  storeName = $store
  password  = $pass
}
$merchantId = [int64]$reg.Data.merchantId
$mToken = [string]$reg.Data.accessToken
Assert-True ($merchantId -gt 0 -and $mToken.Length -gt 10) "merchant register id=$merchantId"

$mAuth = @{ Authorization = "Bearer $mToken" }

# 2) 免费额度 2 宝宝
$babies = @()
foreach ($i in 1..2) {
  $b = Invoke-Api -Method POST -Path "/device/merchant/api/baby/create" -Headers $mAuth -Body @{
    babyName = "宝宝$i"
    birthday = ([DateTimeOffset]::UtcNow.AddYears(-1).ToUnixTimeSeconds())
    sex      = 1
  }
  $babies += [string]$b.Data.deviceNo
  Assert-True ($babies[-1].Length -gt 0) "create baby $i deviceNo=$($babies[-1])"
}

# 3) 第 3 个应失败（未授期）
$thirdFail = $false
try {
  Invoke-Api -Method POST -Path "/device/merchant/api/baby/create" -Headers $mAuth -Body @{
    babyName = "宝宝3"
    birthday = ([DateTimeOffset]::UtcNow.AddYears(-1).ToUnixTimeSeconds())
    sex      = 0
  } | Out-Null
} catch {
  $thirdFail = $true
  Write-Host "OK  third baby rejected before grant: $($_.Exception.Message)"
}
Assert-True $thirdFail "third baby blocked without entitlement"

# 4) staff + 分发
$staff = Invoke-Api -Method POST -Path "/device/merchant/api/staff/create" -Headers $mAuth -Body @{
  account  = $staffAccount
  password = $staffPass
}
$wxId = [int64]$staff.Data.wxId
Assert-True ($wxId -gt 0) "staff wxId=$wxId"

Invoke-Api -Method POST -Path "/device/merchant/api/baby/assign" -Headers $mAuth -Body @{
  wxId     = $wxId
  deviceNo = $babies[0]
  setFocus = $true
} | Out-Null
Write-Host "OK  assigned $($babies[0]) to staff"

$staffLogin = Invoke-Api -Method POST -Path "/device/app/api/username_login" -Body @{
  account  = $staffAccount
  password = $staffPass
}
$staffToken = [string]$staffLogin.Data.accessToken
$focusDn = [string]$staffLogin.Data.deviceNo
Assert-True ($staffToken.Length -gt 10) "staff login"
Assert-True ($focusDn -eq $babies[0]) "staff focus deviceNo=$focusDn"
$sAuth = @{ Authorization = "Bearer $staffToken" }

# 5) 欠费：未授期商户宝宝禁 history create
$now = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$createBlocked = $false
try {
  Invoke-Api -Method POST -Path "/device/history/api/event/add" -Headers $sAuth -Body @{
    deviceNo    = $focusDn
    eventId     = 1
    eventName   = "喂奶"
    eventUnit   = "ml"
    eventNumber = 30
    startTime   = $now
    endTime     = 0
    remark      = "e2e-arrears"
  } | Out-Null
} catch {
  $createBlocked = $true
  Write-Host "OK  history create blocked while arrears: $($_.Exception.Message)"
}
Assert-True $createBlocked "history create blocked on arrears"

# 6) Hub 绝对授期
$adminLogin = Invoke-Api -Method POST -Path "/device/admin/api/login" -Body @{
  username = $AdminUser
  password = $AdminPassword
}
$adminToken = [string]$adminLogin.Data.accessToken
Assert-True ($adminToken.Length -gt 10) "admin login"
$aAuth = @{ Authorization = "Bearer $adminToken" }

$expireAt = [DateTimeOffset]::UtcNow.AddDays(30).ToUnixTimeSeconds()
$grant = Invoke-Api -Method POST -Path "/cash/admin/api/merchant/grant" -Headers $aAuth -Body @{
  merchantId = $merchantId
  expireAt   = $expireAt
}
Assert-True ([int64]$grant.Data.expireAt -eq $expireAt) "Hub grant expireAt=$expireAt"

# 7) 第 3 宝宝
$b3 = Invoke-Api -Method POST -Path "/device/merchant/api/baby/create" -Headers $mAuth -Body @{
  babyName = "宝宝3"
  birthday = ([DateTimeOffset]::UtcNow.AddYears(-1).ToUnixTimeSeconds())
  sex      = 0
}
$babies += [string]$b3.Data.deviceNo
Assert-True ($babies[2].Length -gt 0) "third baby after grant=$($babies[2])"

# 8) create + end 历史
$now2 = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$add = Invoke-Api -Method POST -Path "/device/history/api/event/add" -Headers $sAuth -Body @{
  deviceNo    = $focusDn
  eventId     = 1
  eventName   = "喂奶"
  eventUnit   = "ml"
  eventNumber = 40
  startTime   = $now2
  endTime     = 0
  remark      = "e2e-ok"
}
$hid = [int64]$add.Data.id
Assert-True ($hid -gt 0) "history create id=$hid"

Invoke-Api -Method POST -Path "/device/history/api/event/end-latest" -Headers $sAuth -Body @{
  deviceNo  = $focusDn
  eventId   = 1
  eventName = "喂奶"
} | Out-Null
Write-Host "OK  history end-latest"

# 9.3 subscribe + simulate-paid（需 CASH_PAYMENT_DEV_BYPASS=1）
$sub = Invoke-Api -Method POST -Path "/cash/merchant/api/subscribe" -Headers $mAuth -Body @{}
$orderNo = [string]$sub.Data.orderNo
Assert-True ($orderNo.Length -gt 0) "subscribe orderNo=$orderNo"

$sim = Invoke-Api -Method POST -Path "/cash/merchant/api/subscribe/simulate-paid" -Headers $mAuth -Body @{
  orderNo = $orderNo
}
Assert-True ([bool]$sim.Data.ok) "simulate-paid ok"

$ent = Invoke-Api -Method GET -Path "/cash/merchant/api/entitlement" -Headers $mAuth
Assert-True ([bool]$ent.Data.entitled) "entitlement entitled after pay"
Assert-True ([int64]$ent.Data.expireAt -ge $expireAt) "expire_at shared/extended: $($ent.Data.expireAt) (hub was $expireAt)"
Write-Host "OK  entitlement source=$($ent.Data.source) expireAt=$($ent.Data.expireAt)"

Write-Host ""
Write-Host "=== E2E PASSED merchantId=$merchantId store=$store ==="
Write-Host "Note: SSE 局部刷新请在浏览器打开 /merchant 手工确认一次 create 推送。"

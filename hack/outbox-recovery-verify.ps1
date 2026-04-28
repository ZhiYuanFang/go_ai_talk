param(
    [string]$MySqlHost = "127.0.0.1",
    [int]$MySqlPort = 3306,
    [string]$MySqlUser = "root",
    [string]$MySqlPassword = "",
    [string]$MySqlDatabase = "ai_voice_history",
    [switch]$ShowPendingOnly,
    [switch]$ResetFailedToPending
)

$ErrorActionPreference = "Stop"

function Get-MySqlCli {
    $cmd = Get-Command mysql -ErrorAction SilentlyContinue
    if (-not $cmd) {
        throw "未找到 mysql 命令，请先安装 MySQL 客户端并加入 PATH。"
    }
    return $cmd.Source
}

function Invoke-MySqlQuery {
    param([string]$Sql)
    $mysql = Get-MySqlCli
    $args = @(
        "-h", $MySqlHost,
        "-P", "$MySqlPort",
        "-u", $MySqlUser,
        "-D", $MySqlDatabase,
        "-e", $Sql
    )
    if ($MySqlPassword -ne "") {
        $args += "-p$MySqlPassword"
    }
    & $mysql @args
}

Write-Host "检查 domain_outbox 状态..."
Invoke-MySqlQuery "SELECT status, COUNT(*) AS cnt FROM domain_outbox GROUP BY status ORDER BY status;"

if ($ShowPendingOnly) {
    Write-Host ""
    Write-Host "展示 pending/failed 事件（最多 30 条）..."
    Invoke-MySqlQuery "SELECT id,event_id,routing_key,status,attempts,last_error,created_at,updated_at FROM domain_outbox WHERE status IN ('pending','failed') ORDER BY id DESC LIMIT 30;"
}

if ($ResetFailedToPending) {
    Write-Host ""
    Write-Host "将 failed 事件重置为 pending（用于恢复后重放）..."
    Invoke-MySqlQuery "UPDATE domain_outbox SET status='pending', last_error='', updated_at=NOW() WHERE status='failed';"
    Invoke-MySqlQuery "SELECT ROW_COUNT() AS reset_rows;"
}

Write-Host ""
Write-Host "完成。"

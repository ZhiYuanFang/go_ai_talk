# RabbitMQ 基线：声明 voice.events（topic）及 5 个队列与绑定。
# - 与 docs/runbooks/rabbitmq-local.md、hack/rabbitmq-init.sh 语义一致。
# - 在「全新」Broker 上可重复执行；若绑定已存在，部分 Rabbit 版本可能对重复 POST binding 返回 4xx，需删队列或换 vhost 后重跑。
# - 宿主机执行时 ApiBase 用 127.0.0.1:15672；在远端机器请改为可达的管理 API 地址。
param(
    [string]$ComposeFile = "manifest/docker/docker-compose.rabbitmq.yml",
    [string]$ApiBase = "http://127.0.0.1:15672/api",
    [string]$User = "guest",
    [string]$Password = "guest",
    [switch]$PrintOnly
)

$ErrorActionPreference = "Stop"

function Invoke-RabbitApi {
    param(
        [string]$Method,
        [string]$Path,
        [object]$Body = $null
    )
    $uri = $ApiBase.TrimEnd("/") + $Path
    $pair = "{0}:{1}" -f $User, $Password
    $token = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes($pair))
    $headers = @{ Authorization = "Basic $token" }
    if ($Body -ne $null) {
        $json = $Body | ConvertTo-Json -Compress
        return Invoke-RestMethod -Method $Method -Uri $uri -Headers $headers -Body $json -ContentType "application/json"
    }
    return Invoke-RestMethod -Method $Method -Uri $uri -Headers $headers
}

if (-not $PrintOnly) {
    Write-Host "Starting RabbitMQ..."
    docker compose -f $ComposeFile up -d
    Start-Sleep -Seconds 6
}

Write-Host "Declaring exchange and queues..."

$exchange = "/exchanges/%2F/voice.events"
$queues = @(
    @{ Name = "voice.task.requested.q"; RoutingKey = "voice.task.requested" },
    @{ Name = "voice.task.completed.q"; RoutingKey = "voice.task.completed" },
    @{ Name = "voice.task.failed.q"; RoutingKey = "voice.task.failed" },
    @{ Name = "notify.events.q"; RoutingKey = "notify.*" },
    @{ Name = "history.events.q"; RoutingKey = "history.#" },
    @{ Name = "ucg.post.created.q"; RoutingKey = "ucg.post.created" },
    @{ Name = "ucg.comment.created.q"; RoutingKey = "ucg.comment.created" },
    @{ Name = "ucg.profile.patch.submitted.q"; RoutingKey = "ucg.profile.patch.submitted" },
    @{ Name = "ucg.chat.msg.created.q"; RoutingKey = "ucg.chat.msg.created" },
    @{ Name = "ucg.recommend.score.q"; RoutingKey = "ucg.post.published" },
    @{ Name = "ucg.recommend.score.q"; RoutingKey = "ucg.post.unpublished" },
    @{ Name = "ucg.recommend.score.q"; RoutingKey = "ucg.post.liked" },
    @{ Name = "ucg.recommend.score.q"; RoutingKey = "ucg.post.unliked" },
    @{ Name = "ucg.recommend.score.q"; RoutingKey = "ucg.comment.published" },
    @{ Name = "ucg.recommend.score.q"; RoutingKey = "ucg.comment.removed" }
)

if ($PrintOnly) {
    Write-Host "PrintOnly mode: skip actual API calls."
} else {
    Invoke-RabbitApi -Method Put -Path $exchange -Body @{ type = "topic"; durable = $true; auto_delete = $false }
    foreach ($q in $queues) {
        $queuePath = "/queues/%2F/$($q.Name)"
        Invoke-RabbitApi -Method Put -Path $queuePath -Body @{ durable = $true; auto_delete = $false; arguments = @{} }
        $bindPath = "/bindings/%2F/e/voice.events/q/$($q.Name)"
        Invoke-RabbitApi -Method Post -Path $bindPath -Body @{ routing_key = $q.RoutingKey; arguments = @{} }
    }
}

Write-Host ""
Write-Host "RabbitMQ baseline initialized."
Write-Host "Exchange: voice.events (topic)"
Write-Host "Queues:"
foreach ($q in $queues) {
    Write-Host (" - " + $q.Name + " <= " + $q.RoutingKey)
}
Write-Host ""
Write-Host "Verify in UI: http://127.0.0.1:15672 (guest/guest)"

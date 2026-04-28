param(
    [string]$ComposeFile = "manifest/docker/docker-compose.redis-cluster.yml"
)

$ErrorActionPreference = "Stop"

Write-Host "Starting redis cluster nodes..."
docker compose -f $ComposeFile up -d

Write-Host "Waiting for nodes to become ready..."
Start-Sleep -Seconds 5

$nodes = @(
    "redis-node-1:7001",
    "redis-node-2:7002",
    "redis-node-3:7003",
    "redis-node-4:7004",
    "redis-node-5:7005",
    "redis-node-6:7006"
)

$createCmd = @(
    "redis-cli",
    "--cluster", "create",
    $nodes,
    "--cluster-replicas", "1",
    "--cluster-yes"
)

Write-Host "Creating redis cluster (3 masters + 3 replicas)..."
docker compose -f $ComposeFile exec -T redis-node-1 $createCmd

Write-Host ""
Write-Host "Redis cluster initialized."
Write-Host "Validate with:"
Write-Host "  docker compose -f $ComposeFile exec -T redis-node-1 redis-cli -p 7001 cluster info"
Write-Host "  docker compose -f $ComposeFile exec -T redis-node-1 redis-cli -p 7001 cluster nodes"

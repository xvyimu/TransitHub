# Empty SQLite dry-run for migrations/main (Phase1 WP-S)
# Does not touch production DSN.

[CmdletBinding()]
param(
  [string]$MigrationsPath = ""
)

$ErrorActionPreference = "Stop"
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path.TrimEnd("\")
if ([string]::IsNullOrWhiteSpace($MigrationsPath)) {
  $MigrationsPath = Join-Path $repoRoot "migrations\main"
}
$MigrationsPath = (Resolve-Path $MigrationsPath).Path

$tmpDir = Join-Path $repoRoot ".tmp"
New-Item -ItemType Directory -Force -Path $tmpDir | Out-Null
$dbFile = Join-Path $tmpDir ("migrate-dry-{0}.db" -f (Get-Date -Format "yyyyMMddHHmmss"))
if (Test-Path $dbFile) { Remove-Item -Force $dbFile }
$dbURL = "sqlite://" + (($dbFile -replace '\\', '/'))

Write-Host "migrations: $MigrationsPath"
Write-Host "database:   $dbURL"

Push-Location $repoRoot
try {
  go run ./cmd/dbmigrate -path $MigrationsPath -database $dbURL up
  if ($LASTEXITCODE -ne 0) { throw "up failed: $LASTEXITCODE" }
  go run ./cmd/dbmigrate -path $MigrationsPath -database $dbURL version
  if ($LASTEXITCODE -ne 0) { throw "version failed: $LASTEXITCODE" }
  Write-Host "PASS sql-migrate-dry-run ($dbFile)"
  exit 0
} finally {
  Pop-Location
}

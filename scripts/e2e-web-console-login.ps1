# Login e2e for web-console API subset (Phase1)
# Usage:
#   $env:TH_E2E_USER='...'; $env:TH_E2E_PASS='...'
#   pwsh -File scripts/e2e-web-console-login.ps1
# Defaults: root / 123456 (only valid on empty DB after createRootAccountIfNeed)
#
# Prefer scripts/w4-d7-nonprod-verify.ps1 for G2/G3 gates: it refuses silent
# root defaults and exits 10 when TH_E2E_* missing (no fake green).
# Exit map + failure modes: docs/ops/th-day-e2e-harness-2026-07-24.md

param(
  [string]$ApiBase = $(if ($env:TH_API_BASE) { $env:TH_API_BASE } else { 'http://127.0.0.1:3000' }),
  [string]$ViteBase = $(if ($env:TH_VITE_BASE) { $env:TH_VITE_BASE } else { 'http://127.0.0.1:5173' }),
  [switch]$SkipVite
)

$ErrorActionPreference = 'Stop'
$user = if ($env:TH_E2E_USER) { $env:TH_E2E_USER } else { 'root' }
$pass = if ($env:TH_E2E_PASS) { $env:TH_E2E_PASS } else { '123456' }
if (-not $env:TH_E2E_USER -or -not $env:TH_E2E_PASS) {
  Write-Host "WARN using default root/123456 (TH_E2E_* incomplete). Prefer w4-d7-nonprod-verify.ps1 (exit 10, no silent default)."
}

function Write-Step($msg) { Write-Host "==> $msg" }

try {
  Write-Step "healthz $ApiBase"
  $hz = Invoke-WebRequest -Uri "$ApiBase/healthz" -UseBasicParsing -TimeoutSec 5
  if ($hz.StatusCode -ne 200) { throw "healthz status $($hz.StatusCode)" }
} catch {
  Write-Host "ERR backend unreachable: $_"
  exit 3
}

Write-Step "POST /api/user/login as $user"
$body = @{ username = $user; password = $pass } | ConvertTo-Json
try {
  $login = Invoke-WebRequest -Uri "$ApiBase/api/user/login" -Method POST -Body $body -ContentType 'application/json; charset=utf-8' -SessionVariable sess -UseBasicParsing -TimeoutSec 10
} catch {
  Write-Host "ERR login request: $_"
  exit 1
}
$loginJson = $login.Content | ConvertFrom-Json
if (-not $loginJson.success) {
  Write-Host "ERR login failed: $($login.Content)"
  exit 1
}
Write-Host "login OK id=$($loginJson.data.id) role=$($loginJson.data.role)"

$uid = $loginJson.data.id
if (-not $uid) {
  Write-Host "ERR login response missing data.id (needed for New-Api-User)"
  exit 1
}

Write-Step "GET /api/user/self (New-Api-User=$uid)"
try {
  $self = Invoke-WebRequest -Uri "$ApiBase/api/user/self" -WebSession $sess -Headers @{ 'New-Api-User' = "$uid" } -UseBasicParsing -TimeoutSec 10
} catch {
  Write-Host "ERR self: $_"
  exit 2
}
$selfJson = $self.Content | ConvertFrom-Json
if (-not $selfJson.success) {
  Write-Host "ERR self payload: $($self.Content)"
  exit 2
}
Write-Host "self OK username=$($selfJson.data.username)"

Write-Step "GET /api/status (session optional)"
$st = Invoke-WebRequest -Uri "$ApiBase/api/status" -WebSession $sess -UseBasicParsing -TimeoutSec 10
Write-Host "status HTTP $($st.StatusCode)"

if (-not $SkipVite) {
  try {
    Write-Step "Vite proxy $ViteBase/api/status"
    $vs = Invoke-WebRequest -Uri "$ViteBase/api/status" -UseBasicParsing -TimeoutSec 3
    Write-Host "vite proxy HTTP $($vs.StatusCode)"
  } catch {
    Write-Host "WARN vite proxy skip: $_"
  }
}

Write-Step "PASS e2e-web-console-login"
exit 0

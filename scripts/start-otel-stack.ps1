[CmdletBinding()]
<#
.SYNOPSIS
  Start the LOCAL-ONLY OTEL observability stack for new-api / TransitHub.

.DESCRIPTION
  OTEL Phase 1 (traces) + Phase 2 (logs correlation). Launches, as separate
  local single binaries (R2 LOCAL-ONLY / R3 single-binary discipline):

    - otelcol-contrib : receives OTLP/gRPC traces from new-api on :4317,
                        exports traces to Jaeger and TAILS the on-disk log
                        file (filelog receiver) into Loki.
    - jaeger          : all-in-one traces backend, UI on :16686.
    - loki            : logs backend, API on :3100.

  The new-api binary itself gains NO new module dependency in Phase 2 — it only
  stamps otel_trace_id / otel_span_id into its existing log file, which the
  Collector tails. So this stack is entirely optional and side-car: if it is
  down, new-api keeps serving (traces fail-open, logs still written to disk).

  Nothing here binds to a public interface. Do not port-forward these.

.PARAMETER LogFile
  Path to the new-api log file the Collector should tail. Defaults to the
  newest logs/oneapi-*.log under the repo root.

.PARAMETER OtelcolPath / JaegerPath / LokiPath
  Paths to the respective binaries. Default to the command name on PATH.

.EXAMPLE
  pwsh scripts/start-otel-stack.ps1
  # then, in the new-api environment:
  #   $env:OTEL_TRACES_ENABLED = 'true'
  #   $env:OTEL_LOGS_ENABLED   = 'true'
  #   restart new-api
#>
param(
  [string]$LogFile = "",
  [string]$OtelcolPath = "otelcol-contrib",
  [string]$JaegerPath = "jaeger-all-in-one",
  [string]$LokiPath = "loki",
  [switch]$SkipJaeger,
  [switch]$SkipLoki
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path.TrimEnd("\")
$otelDir = Join-Path $repoRoot "deploy\otel"
$colConfig = Join-Path $otelDir "otelcol-config.yaml"
$lokiConfig = Join-Path $otelDir "loki-config.yaml"

foreach ($f in @($colConfig, $lokiConfig)) {
  if (-not (Test-Path -LiteralPath $f)) { throw "Missing config: $f" }
}

# Resolve the log file to tail: explicit -LogFile, else newest oneapi-*.log.
if ([string]::IsNullOrWhiteSpace($LogFile)) {
  $logsDir = Join-Path $repoRoot "logs"
  if (Test-Path -LiteralPath $logsDir) {
    $newest = Get-ChildItem -LiteralPath $logsDir -Filter "oneapi-*.log" -ErrorAction SilentlyContinue |
      Sort-Object LastWriteTime -Descending | Select-Object -First 1
    if ($newest) { $LogFile = $newest.FullName }
  }
}
if ([string]::IsNullOrWhiteSpace($LogFile)) {
  Write-Warning "No new-api log file found under $repoRoot\logs. The Collector filelog receiver will wait for one to appear."
  $LogFile = Join-Path $repoRoot "logs\oneapi-current.log"
}

# The Collector reads the log path from an env var (see otelcol-config.yaml
# 'include: [ ${env:NEWAPI_LOG_FILE} ]'), so we never rewrite the config.
$env:NEWAPI_LOG_FILE = $LogFile
Write-Host "OTEL stack (LOCAL-ONLY):"
Write-Host "  log file to tail : $LogFile"
Write-Host "  collector config : $colConfig"
Write-Host "  loki config      : $lokiConfig"

function Assert-OnPath([string]$exe) {
  if (-not (Get-Command $exe -ErrorAction SilentlyContinue)) {
    throw "Binary '$exe' not found on PATH. Install it (single binary) or pass an explicit -*Path. This script does not download anything."
  }
}

$procs = @()

if (-not $SkipJaeger) {
  Assert-OnPath $JaegerPath
  # Jaeger all-in-one: OTLP receivers on default 4317/4318 conflict with the
  # Collector, so we point the Collector at Jaeger's OTLP port and let Jaeger
  # own the UI only. Here we run Jaeger with collector OTLP disabled by using
  # its default UI + gRPC query; the Collector exports to jaeger via OTLP 14317.
  $jaegerArgs = @("--collector.otlp.grpc.host-port=127.0.0.1:14317",
                  "--query.http-server.host-port=127.0.0.1:16686")
  Write-Host "starting jaeger-all-in-one (UI http://127.0.0.1:16686)..."
  $procs += Start-Process -FilePath $JaegerPath -ArgumentList $jaegerArgs -PassThru -NoNewWindow
}

if (-not $SkipLoki) {
  Assert-OnPath $LokiPath
  Write-Host "starting loki (API http://127.0.0.1:3100)..."
  $procs += Start-Process -FilePath $LokiPath -ArgumentList @("-config.file=$lokiConfig") -PassThru -NoNewWindow
}

Assert-OnPath $OtelcolPath
Write-Host "starting otelcol-contrib (OTLP/gRPC 127.0.0.1:4317)..."
$procs += Start-Process -FilePath $OtelcolPath -ArgumentList @("--config=$colConfig") -PassThru -NoNewWindow

Write-Host ""
Write-Host "Stack started. PIDs: $($procs.Id -join ', ')"
Write-Host "To enable emission from new-api, set in its environment and restart:"
Write-Host "  OTEL_TRACES_ENABLED=true          # export spans to Collector :4317"
Write-Host "  OTEL_LOGS_ENABLED=true            # stamp otel_trace_id/otel_span_id on log lines"
Write-Host "  OTEL_EXPORTER_OTLP_ENDPOINT=127.0.0.1:4317   (default)"
Write-Host "  Both gates are independent and default false."
Write-Host ""
Write-Host "Stop the stack with: Get-Process otelcol-contrib,jaeger-all-in-one,loki | Stop-Process"

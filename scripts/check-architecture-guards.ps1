<#
.SYNOPSIS
Rejects newly introduced violations of the backend JSON and layering rules.

.DESCRIPTION
Existing debt is intentionally not made a merge blocker. The guard examines
only additions relative to the merge base and fails when production controller
code starts using model.DB directly, or when code using encoding/json directly
marshals/unmarshals JSON instead of common/json.go.
#>
[CmdletBinding()]
param(
    [string]$BaseRef = 'origin/main'
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$repoRoot = (& git rev-parse --show-toplevel).Trim()
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($repoRoot)) {
    throw 'Run this guard from inside a Git worktree.'
}
Set-Location $repoRoot

& git rev-parse --verify --quiet "$BaseRef^{commit}" | Out-Null
if ($LASTEXITCODE -ne 0) {
    throw "Base ref '$BaseRef' does not resolve to a commit. Fetch it or pass -BaseRef <commit>."
}

$mergeBase = (& git merge-base $BaseRef HEAD).Trim()
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($mergeBase)) {
    throw "Cannot find a merge base between '$BaseRef' and HEAD."
}

$patch = @(& git diff --no-ext-diff --unified=0 --diff-filter=AM $mergeBase -- '*.go')
if ($LASTEXITCODE -ne 0) {
    throw 'Unable to inspect the Go diff.'
}

$jsonAliasesByFile = @{}
$violations = [System.Collections.Generic.List[string]]::new()
$currentFile = $null

foreach ($line in $patch) {
    if ($line -like '+++ b/*') {
        $currentFile = $line.Substring(6).Replace('/', [IO.Path]::DirectorySeparatorChar)
        continue
    }

    if ($null -eq $currentFile -or -not $line.StartsWith('+') -or $line.StartsWith('+++')) {
        continue
    }

    $addedCode = $line.Substring(1)
    if (-not $jsonAliasesByFile.ContainsKey($currentFile)) {
        $filePath = Join-Path $repoRoot $currentFile
        $aliases = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal)
        $dotImport = $false
        if (Test-Path -LiteralPath $filePath) {
            $contents = Get-Content -LiteralPath $filePath -Raw
            foreach ($import in [regex]::Matches($contents, '(?m)^\s*(?:(?<alias>[A-Za-z_]\w*|\.|_)\s+)?"encoding/json"\s*$')) {
                $alias = $import.Groups['alias'].Value
                if ([string]::IsNullOrEmpty($alias)) {
                    [void]$aliases.Add('json')
                } elseif ($alias -eq '.') {
                    $dotImport = $true
                } elseif ($alias -ne '_') {
                    [void]$aliases.Add($alias)
                }
            }
        }
        $jsonAliasesByFile[$currentFile] = @{ Aliases = $aliases; DotImport = $dotImport }
    }

    $jsonImport = $jsonAliasesByFile[$currentFile]
    $jsonCall = $false
    foreach ($alias in $jsonImport.Aliases) {
        if ($addedCode -match ("\b" + [regex]::Escape($alias) + '\.(?:Marshal|Unmarshal|NewEncoder|NewDecoder)\s*\(')) {
            $jsonCall = $true
            break
        }
    }
    if (-not $jsonCall -and $jsonImport.DotImport -and $addedCode -match '(?<!\.)\b(?:Marshal|Unmarshal|NewEncoder|NewDecoder)\s*\(') {
        $jsonCall = $true
    }
    if ($jsonCall) {
        $violations.Add("${currentFile}: direct encoding/json marshal/unmarshal call added: $addedCode")
    }

    if ($currentFile -like "controller$([IO.Path]::DirectorySeparatorChar)*" -and
        $currentFile -notlike '*_test.go' -and
        $addedCode -match '\bmodel\.DB\b') {
        $violations.Add("${currentFile}: controller must not access model.DB directly: $addedCode")
    }
}

if ($violations.Count -gt 0) {
    Write-Error "Architecture guard failed against merge base ${mergeBase}:`n$($violations -join "`n")"
    exit 1
}

Write-Host "Architecture guard passed (base $mergeBase; only added Go lines checked)."

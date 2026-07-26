<#
.SYNOPSIS
Blocks new features from sneaking into the legacy web/default frontend tree.

.DESCRIPTION
web/default/ is the D7-before production default and the <=5min rollback surface
(see docs/legacy-frontend-gate.md, docs/PROJECT.md 2.2). It must NOT be deleted,
but new visual/functional work belongs in web-console/ (Vue) until cutover.

This guard implements the d-leg assertions (docs/ops/d-leg-legacy-guard.md 2):

  G-LEG-1 (blocking): a PR whose diff touches web/default/** must carry a
    `LEGACY-HOTFIX:` marker plus an incident/regression link in the PR body,
    unless every changed path is in the build/tooling allowlist below.

  G-LEG-2 (soft warning): new `feat(web/default)` / `feat(web):` commits that
    land on the legacy tree emit ::warning:: (noise observation phase, not
    blocking). Escape: commit body carries a LEGACY-HOTFIX marker.

  G-LEG-4 (blocking): a single PR that both adds web/default/src/features/**
    and web-console/src/** is a suspected React+Vue dual-write and fails
    (gate Resolution B: no long-lived dual-write).

G-LEG-3 (allowlist) — the ONLY categories that justify a legit legacy change,
i.e. the reasons a LEGACY-HOTFIX marker may be applied, or that populate the
pure build/tooling path allowlist below:
  - Security fixes (XSS / auth bypass / a CVE in this tree's dependencies)
  - Production regression fixes (attach an incident link)
  - build/tooling fixes that keep the embed/rollback build green
  - i18n typos already broken in production

Existing debt (a72c558c V2 Atelier chrome) is accepted pre-guard and NOT
retroactively blocked; this guard only inspects changes since the merge base.
#>
[CmdletBinding()]
param(
    [string]$BaseRef = 'origin/main',
    [string]$PrBody = ''
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

function Get-ChangedFiles {
    param([string[]]$Pathspec)
    $out = @(& git diff --name-only --diff-filter=ACMR $mergeBase -- @Pathspec)
    if ($LASTEXITCODE -ne 0) {
        throw "Unable to inspect the diff for $($Pathspec -join ', ')."
    }
    return @($out | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
}

# G-LEG-3 pure build/tooling allowlist: files under web/default that only keep
# the embed/rollback build green and never carry product features. Changes
# limited to these paths do not require a LEGACY-HOTFIX marker. Keep this list
# conservative; anything feature-adjacent must go through the marker instead.
$allowlistPatterns = @(
    'web/default/bun.lock',
    'web/default/bun.lockb'
)

$violations = [System.Collections.Generic.List[string]]::new()

# Plain directory pathspecs match recursively in Git (prefix match), which is
# more predictable than fnmatch `**` globs across Git versions.
# --- G-LEG-1: web/default touched without LEGACY-HOTFIX marker + link ---
$defaultChanged = @(Get-ChangedFiles -Pathspec 'web/default/')
if ($defaultChanged.Count -gt 0) {
    $nonAllowlisted = @($defaultChanged | Where-Object {
        $file = $_
        -not ($allowlistPatterns | Where-Object { $file -like $_ })
    })

    if ($nonAllowlisted.Count -gt 0) {
        $hasMarker = $PrBody -match '(?i)LEGACY-HOTFIX'
        $hasLink = $PrBody -match '(?i)https?://' -or $PrBody -match '#\d+'
        if (-not $hasMarker -or -not $hasLink) {
            $reason = if (-not $hasMarker) {
                'missing LEGACY-HOTFIX marker'
            } else {
                'missing incident/regression link'
            }
            $violations.Add(
                "G-LEG-1: web/default touched ($reason). Non-allowlisted files:`n  " +
                ($nonAllowlisted -join "`n  ") +
                "`n  Add a `LEGACY-HOTFIX:` marker plus an incident/regression link to the PR body, " +
                "or move new work into web-console/ (Vue)."
            )
        }
    }
}

# --- G-LEG-4: React+Vue dual-write in one PR ---
$featuresChanged = @(Get-ChangedFiles -Pathspec 'web/default/src/features/')
$consoleChanged = @(Get-ChangedFiles -Pathspec 'web-console/src/')
if ($featuresChanged.Count -gt 0 -and $consoleChanged.Count -gt 0) {
    $violations.Add(
        "G-LEG-4: possible React+Vue dual-write in one PR (gate Resolution B forbids long-lived dual-write).`n" +
        "  web/default/src/features/**:`n    " + ($featuresChanged -join "`n    ") + "`n" +
        "  web-console/src/**:`n    " + ($consoleChanged -join "`n    ") + "`n" +
        "  Split the change so a capability lands in only one tree."
    )
}

# --- G-LEG-2: soft warning on new feat(web/default) / feat(web) commits ---
$commitLines = @(& git log --format='%H%x1f%s%x1f%b%x1e' "$mergeBase..HEAD" -- 'web/default/**')
if ($LASTEXITCODE -ne 0) {
    throw 'Unable to inspect the commit log for web/default.'
}
$commits = ($commitLines -join "`n") -split "`u{1e}" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
foreach ($commit in $commits) {
    $parts = $commit.TrimStart("`n") -split "`u{1f}"
    if ($parts.Count -lt 2) { continue }
    $hash = $parts[0].Substring(0, [Math]::Min(9, $parts[0].Length))
    $subject = $parts[1]
    $body = if ($parts.Count -ge 3) { $parts[2] } else { '' }
    if ($subject -match '^\s*feat\(web(/default)?\)' ) {
        if ($body -match '(?i)LEGACY-HOTFIX') { continue }
        Write-Host "::warning::G-LEG-2: new feat() on legacy web/default needs a cutover exception: $hash $subject"
    }
}

if ($violations.Count -gt 0) {
    $joined = $violations -join "`n`n"
    foreach ($line in ($joined -split "`n")) {
        Write-Host "::error::$line"
    }
    Write-Error "Legacy-guard failed against merge base ${mergeBase}:`n$joined"
    exit 1
}

Write-Host "Legacy-guard passed (base $mergeBase; G-LEG-1/2/4 checked against changes since merge base)."

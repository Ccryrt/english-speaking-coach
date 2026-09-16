# Agent entry point. No language runtime or separate API key installation.
$ErrorActionPreference = 'Stop'
$coachRoot = Split-Path $PSScriptRoot -Parent
$coachVersion = (Get-Content (Join-Path $coachRoot 'runtime-version.txt') -Raw).Trim()
if ($coachVersion -notmatch '^v[0-9]+\.[0-9]+\.[0-9]+$') { throw 'Invalid pinned runtime version' }
$coachArch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
if ($coachArch -eq 'x64') { $coachArch = 'amd64' }
if ($coachArch -notin @('amd64','arm64')) { throw 'Unsupported CPU architecture' }
$coachName = "english-coach_${coachVersion}_windows_${coachArch}"
$coachBinDir = Join-Path $coachRoot 'bin'
$coachBinary = Join-Path $coachBinDir "$coachName.exe"
$coachReceipt = "$coachBinary.sha256"
if (-not (Test-Path $coachReceipt)) { throw 'Runtime receipt missing; rebuild/reinstall the plugin.' }
if (Test-Path $coachBinary) {
    $coachCachedExpected = (Get-Content -LiteralPath $coachReceipt -Raw).Trim()
    if ((Get-FileHash -Algorithm SHA256 -LiteralPath $coachBinary).Hash.ToLowerInvariant() -ne $coachCachedExpected) {
        throw 'Cached runtime checksum mismatch; nothing executed. Agent must restore the verified pinned release.'
    }
}
if (-not (Test-Path $coachBinary)) {
    throw 'Combined runtime missing. Rebuild/reinstall the English + Japanese plugin.'
}

$env:ENGLISH_COACH_SKILL_ROOT = $coachRoot
& $coachBinary @args
exit $LASTEXITCODE

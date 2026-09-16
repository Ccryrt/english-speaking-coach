$ErrorActionPreference = "Stop"
$coachRoot = Split-Path $PSScriptRoot -Parent
$coachBinary = Join-Path $coachRoot "bin/japanese-coach.exe"
if (!(Test-Path $coachBinary) -or !(Test-Path "$coachBinary.sha256")) {
    throw "Japanese runtime missing. Build this plugin with scripts/build.py on Windows."
}
$expected = (Get-Content "$coachBinary.sha256" -Raw).Trim()
if ((Get-FileHash $coachBinary -Algorithm SHA256).Hash.ToLowerInvariant() -ne $expected) {
    throw "Japanese runtime checksum mismatch; rebuild the Japanese plugin."
}
$env:JAPANESE_COACH_SKILL_ROOT = $coachRoot
& $coachBinary @args
exit $LASTEXITCODE

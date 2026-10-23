param([switch]$CompleteExisting)
$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$envPath = Join-Path $root '.env'
$exists = Test-Path -LiteralPath $envPath
if ($exists -and -not $CompleteExisting) { throw '.env exists. Use -CompleteExisting to fill missing JWT_SECRET or replace a short API_KEY; existing database settings will be preserved.' }
function New-Secret {
    $bytes = New-Object byte[] 32
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
    return [BitConverter]::ToString($bytes).Replace('-', '').ToLowerInvariant()
}
if ($exists) { $text = Get-Content -LiteralPath $envPath -Raw } else { $text = Get-Content -LiteralPath (Join-Path $root '.env.example') -Raw }
$before = $text
$keys = @('API_KEY', 'JWT_SECRET')
if (-not $exists) { $keys = @('DB_PASSWORD', 'API_KEY', 'JWT_SECRET') }
foreach ($key in $keys) {
    $match = [regex]::Match($text, "(?m)^$key=([^\r\n]*)")
    if ($match.Success -and $match.Groups[1].Value.Trim().Length -ge 32) { continue }
    $value = New-Secret
    if ($match.Success) { $text = [regex]::Replace($text, "(?m)^$key=[^\r\n]*", "$key=$value") }
    else { $text = $text.TrimEnd() + "`n$key=$value`n" }
}
if ($exists -and $text -ne $before) { Copy-Item -LiteralPath $envPath -Destination ($envPath + '.backup.' + (Get-Date -Format 'yyyyMMddHHmmss')) }
[IO.File]::WriteAllText($envPath, $text)
Write-Output 'Environment prepared. Existing DB settings were preserved. Secret values were not printed.'

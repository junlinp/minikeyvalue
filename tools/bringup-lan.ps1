# Run master + 3 volumes entirely on D:\ (no C: cache/temp).
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$Data = Join-Path $Root "data"
$Cache = Join-Path $Root ".cache"
New-Item -ItemType Directory -Force -Path "$Data\volume1","$Data\volume2","$Data\volume3","$Data\indexdb","$Data\upload-tmp","$Cache\go-build","$Cache\tmp","$Cache\mod" | Out-Null

$env:GOCACHE = Join-Path $Cache "go-build"
$env:GOTMPDIR = Join-Path $Cache "tmp"
$env:GOMODCACHE = Join-Path $Cache "mod"
$env:TEMP = Join-Path $Cache "tmp"
$env:TMP = Join-Path $Cache "tmp"

$volumed = Join-Path $Root "tools\volumed.exe"
$mkv = Join-Path $Root "mkv.exe"
if (-not (Test-Path $volumed) -or -not (Test-Path $mkv)) {
  Write-Error "Build mkv.exe and tools\volumed.exe first."
}

$env:PORT = "3001"; Start-Process -FilePath $volumed -ArgumentList "$Data\volume1" -WorkingDirectory $Root -WindowStyle Hidden
$env:PORT = "3002"; Start-Process -FilePath $volumed -ArgumentList "$Data\volume2" -WorkingDirectory $Root -WindowStyle Hidden
$env:PORT = "3003"; Start-Process -FilePath $volumed -ArgumentList "$Data\volume3" -WorkingDirectory $Root -WindowStyle Hidden
Start-Sleep -Seconds 1
& $mkv -port 3000 -volumes 127.0.0.1:3001,127.0.0.1:3002,127.0.0.1:3003 -db "$Data\indexdb" -tmpdir "$Data\upload-tmp" -advertise 192.168.0.107 -replicas 3 -v server

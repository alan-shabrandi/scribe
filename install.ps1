$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$exePath = Join-Path $scriptDir "scribe.exe"
$targetDir = "$env:LOCALAPPDATA\Programs\Scribe"

if (!(Test-Path $exePath)) {
    Write-Host "Error: 'scribe.exe' not found in the current directory!" -ForegroundColor Red
    exit 1
}

if (!(Test-Path $targetDir)) {
    New-Item -ItemType Directory -Force -Path $targetDir | Out-Null
}

Copy-Item -Path $exePath -Destination "$targetDir\scribe.exe" -Force

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$targetDir*") {
    $newPath = "$userPath;$targetDir"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "Added Scribe to User PATH." -ForegroundColor Cyan
}

Write-Host "Scribe installed successfully!" -ForegroundColor Green
Write-Host "Please restart your terminal and run scribe from anywhere." -ForegroundColor Yellow
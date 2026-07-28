$targetDir = "$env:LOCALAPPDATA\Programs\Scribe"

if (!(Test-Path $targetDir)) {
    New-Item -ItemType Directory -Force -Path $targetDir | Out-Null
}

if (!(Test-Path ".\scribe.exe")) {
    Write-Host "Error: 'scribe.exe' not found in current directory!" -ForegroundColor Red
    Write-Host "Please build the project first using: go build -o scribe.exe ./cmd/scribe" -ForegroundColor Yellow
    exit 1
}

Copy-Item -Path ".\scribe.exe" -Destination "$targetDir\scribe.exe" -Force

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$targetDir*") {
    $newPath = "$userPath;$targetDir"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "Added Scribe to User PATH." -ForegroundColor Cyan
}

Write-Host "Scribe installed successfully!" -ForegroundColor Green
Write-Host "Please restart your terminal and run scribe from anywhere." -ForegroundColor Yellow
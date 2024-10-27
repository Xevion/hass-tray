# A build & install script for the project's Windows version.
$ErrorActionPreference = "Stop"
$executableName = "door_tray"

if (-Not (Test-Path -Path "./go.mod")) {
    Write-Error "Please run this script from the project's root directory (go.mod not found)."
    exit 1
}

# Build
go build -o "bin/$executableName-temp.exe" -ldflags "-s -w" "./cmd/windows/"
if ($LASTEXITCODE -ne 0) {
    Write-Error "Build failed with exit code $LASTEXITCODE."
    exit 1
}

# Compress
upx "bin/$executableName-temp.exe" -o "bin/$executableName.exe" -5 -f
if ($LASTEXITCODE -ne 0) {
    Write-Error "Compression failed with exit code $LASTEXITCODE."
    exit 1
}
Remove-Item "bin/$executableName-temp.exe"

# Setup service
$serviceName = "DoorTray"

# TODO: Stop old service
# TODO: Install new binary
# TODO: Start & verify service
# TODO: Cleanup, print latest logs
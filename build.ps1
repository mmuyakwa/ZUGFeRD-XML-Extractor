# Build script for ZUGFeRD Extractor (Windows PowerShell)
Write-Host "Building ZUGFeRD XML Extractor..." -ForegroundColor Green

# Create build directory
New-Item -ItemType Directory -Force -Path "build" | Out-Null

# Build for Windows
Write-Host "Building for Windows (amd64)..." -ForegroundColor Yellow
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o "build/zugferd-extractor.exe" ./cmd/zugferd-extractor

# Build for Linux
Write-Host "Building for Linux (amd64)..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o "build/zugferd-extractor-linux" ./cmd/zugferd-extractor

# Reset environment
Remove-Item Env:GOOS
Remove-Item Env:GOARCH

Write-Host "Build complete! Binaries are in the 'build' directory." -ForegroundColor Green
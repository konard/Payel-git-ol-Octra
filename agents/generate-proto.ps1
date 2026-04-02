# Generate proto files for Agents service

Write-Host "Generating proto files for Agents service..." -ForegroundColor Cyan

# Check if protoc is installed
$protoc = Get-Command protoc -ErrorAction SilentlyContinue
if (-not $protoc) {
    Write-Host "protoc not found. Install Protocol Buffers compiler" -ForegroundColor Red
    exit 1
}

# Check if protoc-gen-go is installed
$protoc_gen_go = Get-Command protoc-gen-go -ErrorAction SilentlyContinue
if (-not $protoc_gen_go) {
    Write-Host "protoc-gen-go not found. Installing..." -ForegroundColor Yellow
    go install github.com/golang/protobuf/protoc-gen-go@latest
}

# Check if protoc-gen-go-grpc is installed
$protoc_gen_grpc = Get-Command protoc-gen-go-grpc -ErrorAction SilentlyContinue
if (-not $protoc_gen_grpc) {
    Write-Host "protoc-gen-go-grpc not found. Installing..." -ForegroundColor Yellow
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
}

# Create output directory if it doesn't exist
if (-not (Test-Path "pb")) {
    New-Item -ItemType Directory -Path "pb" | Out-Null
    Write-Host "Created pb directory" -ForegroundColor Green
}

# Generate proto files
Write-Host "Running protoc..." -ForegroundColor Cyan
$protoDir = "proto"
$outputDir = "pb"

protoc `
    --go_out=$outputDir `
    --go-grpc_out=$outputDir `
    --go_opt=paths=source_relative `
    --go-grpc_opt=paths=source_relative `
    -I="$protoDir" `
    "$protoDir/agents.proto"

if ($LASTEXITCODE -eq 0) {
    Write-Host "Proto generation completed successfully!" -ForegroundColor Green
    Write-Host "Generated files:" -ForegroundColor Cyan
    Get-ChildItem -Path $outputDir -Filter "*.pb.go" | ForEach-Object {
        Write-Host "   - $($_.Name)" -ForegroundColor Green
    }
} else {
    Write-Host "Proto generation failed!" -ForegroundColor Red
    exit 1
}

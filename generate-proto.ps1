# Генерация proto кода для всех сервисов crewai

$env:Path = "C:\Users\pasaz\go\bin;" + $env:Path
$env:Path = "C:\Users\pasaz\AppData\Local\Microsoft\WinGet\Packages\Google.Protobuf_Microsoft.Winget.Source_8wekyb3d8bbwe\bin;" + $env:Path

Write-Host "=== Генерация proto для Boss сервиса ===" -ForegroundColor Green
cd C:\Users\pasaz\GolandProjects\crewai\boss
mkdir -p internal\fetcher\grpc\bosspb 2>$null
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=proto proto/gateway-boss.proto
Move-Item -Force *.pb.go internal\fetcher\grpc\bosspb\ 2>$null

Write-Host "=== Генерация proto для Manager сервиса ===" -ForegroundColor Green
cd C:\Users\pasaz\GolandProjects\crewai\manager
mkdir -p internal\fetcher\grpc\managerpb 2>$null
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=proto proto/boss-manager.proto
Move-Item -Force *.pb.go internal\fetcher\grpc\managerpb\ 2>$null

Write-Host "=== Генерация proto для Worker сервиса ===" -ForegroundColor Green
cd C:\Users\pasaz\GolandProjects\crewai\worker
mkdir -p internal\fetcher\grpc\workerpb 2>$null
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=proto proto/manager-worker.proto
Move-Item -Force *.pb.go internal\fetcher\grpc\workerpb\ 2>$null

Write-Host "=== Готово! ===" -ForegroundColor Green

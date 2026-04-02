$env:Path = "C:\Users\pasaz\go\bin;C:\Users\pasaz\AppData\Local\Microsoft\WinGet\Packages\Google.Protobuf_Microsoft.Winget.Source_8wekyb3d8bbwe\bin;" + $env:Path

cd C:\Users\pasaz\GolandProjects\crewai\apigateway

# Генерируем только boss.proto (gateway-boss.proto используется только для совместимости)
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=proto proto/boss.proto

# Удаляем дубликаты из корня (файлы должны быть в internal/fetcher/grpc/boss/bosspb)
if (Test-Path "boss.pb.go") { Move-Item -Force "boss.pb.go" "internal\fetcher\grpc\boss\bosspb\" }
if (Test-Path "boss_grpc.pb.go") { Move-Item -Force "boss_grpc.pb.go" "internal\fetcher\grpc\boss\bosspb\" }
if (Test-Path "gateway-boss.pb.go") { Remove-Item -Force "gateway-boss.pb.go" }
if (Test-Path "gateway-boss_grpc.pb.go") { Remove-Item -Force "gateway-boss_grpc.pb.go" }

Write-Host "Proto сгенерирован!"

$env:Path = "C:\Users\pasaz\go\bin;" + $env:Path
$env:Path = "C:\Users\pasaz\AppData\Local\Microsoft\WinGet\Packages\Google.Protobuf_Microsoft.Winget.Source_8wekyb3d8bbwe\bin;" + $env:Path

cd C:\Users\pasaz\GolandProjects\crewai\boss

# Delete old files
del internal\fetcher\grpc\bosspb\*.pb.go 2>$null
del internal\fetcher\grpc\manager\managerpb\*.pb.go 2>$null
del *.pb.go 2>$null

# Generate boss service proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=proto proto/gateway-boss.proto
move *.pb.go internal\fetcher\grpc\bosspb\ 2>$null

# Generate boss-manager proto (for communication with manager service)
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=proto proto/boss-manager.proto
move *.pb.go internal\fetcher\grpc\manager\managerpb\ 2>$null

Write-Host "Proto generated for boss!"

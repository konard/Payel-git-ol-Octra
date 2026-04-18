$env:Path = "C:\Users\pasaz\go\bin;" + $env:Path
$env:Path = "C:\Users\pasaz\AppData\Local\Microsoft\WinGet\Packages\Google.Protobuf_Microsoft.Winget.Source_8wekyb3d8bbwe\bin;" + $env:Path

cd C:\Users\pasaz\GolandProjects\crewai\worker
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=proto proto/manager-worker.proto

Write-Host "Proto сгенерирован!"

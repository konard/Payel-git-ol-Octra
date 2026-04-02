Set-Location .\apigateway
go run ./cmd/app/main.go
Set-Location .\boss\
go run ./cmd/app/main.go
Set-Location .\manager
go run ./cmd/app/main.go
Set-Location .\worker
go run ./cmd/app/main.go
# Запуск CLIProxyAPI для CrewAI
# Этот скрипт запускает CLIProxyAPI локально для использования с CrewAI

Write-Host "🚀 Запуск CLIProxyAPI для CrewAI..." -ForegroundColor Green

$configPath = Join-Path $PSScriptRoot "cliproxy\config.yaml"
$exePath = Join-Path $PSScriptRoot "cliproxy\cli-proxy-api.exe"

if (-not (Test-Path $exePath)) {
    Write-Host "❌ CLIProxyAPI не найден. Скачиваю..." -ForegroundColor Yellow
    
    # Создаём директорию
    New-Item -ItemType Directory -Force -Path (Join-Path $PSScriptRoot "cliproxy") | Out-Null
    
    # Скачиваем последнюю версию
    $releaseUrl = "https://github.com/router-for-me/CLIProxyAPI/releases/latest/download/CLIProxyAPI_6.9.15_windows_amd64.zip"
    $zipPath = Join-Path $PSScriptRoot "cliproxy\CLIProxyAPI.zip"
    
    Write-Host "⬇️  Скачиваю CLIProxyAPI..." -ForegroundColor Cyan
    Invoke-WebRequest -Uri $releaseUrl -OutFile $zipPath
    
    Write-Host "📦 Распаковываю..." -ForegroundColor Cyan
    Expand-Archive -Path $zipPath -DestinationPath (Join-Path $PSScriptRoot "cliproxy") -Force
    Remove-Item $zipPath -Force
}

if (-not (Test-Path $configPath)) {
    Write-Host "❌ Конфиг не найден!" -ForegroundColor Red
    exit 1
}

Write-Host "✅ CLIProxyAPI найден" -ForegroundColor Green
Write-Host "📋 Конфиг: $configPath" -ForegroundColor Cyan
Write-Host ""
Write-Host "🔗 URL: http://localhost:8317" -ForegroundColor Cyan
Write-Host "🤖 Модель: qwen-code" -ForegroundColor Cyan
Write-Host ""
Write-Host "🔄 Запуск..." -ForegroundColor Green
Write-Host ""

# Запускаем CLIProxyAPI
Start-Process -FilePath $exePath -ArgumentList "--config", $configPath -WindowStyle Normal

Start-Sleep -Seconds 3

# Проверяем что запустился
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8317/v1/models" -Headers @{Authorization = "Bearer cliproxy-dev-key-change-me"}
    Write-Host "✅ CLIProxyAPI запущен успешно!" -ForegroundColor Green
    Write-Host "📊 Доступные модели:" -ForegroundColor Cyan
    $response.data | ForEach-Object { Write-Host "   - $($_.id)" -ForegroundColor White }
} catch {
    Write-Host "⚠️  CLIProxyAPI ещё запускается, подожди несколько секунд..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "💡 Теперь запусти CrewAI:" -ForegroundColor Green
Write-Host "   docker-compose up -d --build"
Write-Host ""
Write-Host "🛑  Чтобы остановить CLIProxyAPI:" -ForegroundColor Yellow
Write-Host "   taskkill /IM cli-proxy-api.exe /F"
Write-Host ""

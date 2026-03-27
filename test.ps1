# Тестирование CrewAI

$baseUrl = "http://localhost:3111"

# Запрос на создание задачи
$body = @{
    username = "pavel"
    taskName = "Личный прокси"
    title = "Требуется создать личный прокси сервер"
    description = "Необходимо создать личный прокси сервер для использования в различных сервисах. Сервер должен быть безопасным и надежным"
    tokens = @(
        "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",
        "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",
        "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"
    )
    meta = @{
        model = "meta-llama/llama-3.3-70b-instruct:free"
        modelUrl = "https://openrouter.ai/api/v1"
        test = "value"
    }
} | ConvertTo-Json -Depth 10

Write-Host "=== Отправка задачи ===" -ForegroundColor Green
$response = Invoke-RestMethod -Uri "$baseUrl/task/create" -Method Post -ContentType "application/json" -Body $body
$response | ConvertTo-Json -Depth 10

Write-Host "`n=== Задача создана ===" -ForegroundColor Green
$taskId = $response.task_id
Write-Host "Task ID: $taskId"

# Проверка статуса
Start-Sleep -Seconds 2
Write-Host "`n=== Проверка статуса ===" -ForegroundColor Green
$statusResponse = Invoke-RestMethod -Uri "$baseUrl/task/status?task_id=$taskId" -Method Get
$statusResponse | ConvertTo-Json -Depth 10

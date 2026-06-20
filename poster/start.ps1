param(
    [Switch]$Install
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

if ($Install) {
    Write-Host "Installing dependencies..." -ForegroundColor Cyan
    python -m uv sync
    Write-Host "Done!" -ForegroundColor Green
    return
}

Write-Host "Starting Octra Poster Bot..." -ForegroundColor Cyan
python -m uv run python -m src.main

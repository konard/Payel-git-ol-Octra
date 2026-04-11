"""
Конфигурация Telegram бота
"""
import os
from dotenv import load_dotenv

# Загружаем .env только если он существует
_env_file = os.path.join(os.path.dirname(__file__), '..', '.env')
if os.path.exists(_env_file):
    load_dotenv(_env_file)

BOT_TOKEN = os.getenv("TG_BOT_TOKEN")
AUTH_API_URL = os.getenv("AUTH_API_URL", "http://localhost:3112")
API_GATEWAY_URL = os.getenv("API_GATEWAY_URL", "ws://localhost:3111")

# Для локальной разработки без Docker
if "localhost" in API_GATEWAY_URL:
    API_GATEWAY_HTTP_URL = os.getenv("API_GATEWAY_HTTP_URL", "http://localhost:3111")
else:
    # Внутри Docker - заменяем ws:// на http://
    API_GATEWAY_HTTP_URL = API_GATEWAY_URL.replace("ws://", "http://").replace("wss://", "https://")

if not BOT_TOKEN:
    raise ValueError(
        "TG_BOT_TOKEN не найден! "
        "Установите переменную окружения TG_BOT_TOKEN или добавьте её в .env файл"
    )

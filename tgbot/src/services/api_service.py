"""
Сервис для работы с API авторизации и задач
"""
import aiohttp
import asyncio
import json
import websockets
from typing import Optional, Dict, Any, AsyncGenerator
from src.config import AUTH_API_URL, API_GATEWAY_URL, API_GATEWAY_HTTP_URL


class APIService:
    """Сервис для работы с API"""
    
    def __init__(self):
        self.auth_url = AUTH_API_URL
        self.gateway_url = API_GATEWAY_URL
        self.gateway_http_url = API_GATEWAY_HTTP_URL
        self._session: Optional[aiohttp.ClientSession] = None
    
    async def get_session(self) -> aiohttp.ClientSession:
        """Получить HTTP сессию"""
        if self._session is None or self._session.closed:
            self._session = aiohttp.ClientSession()
        return self._session
    
    async def close(self):
        """Закрыть сессию"""
        if self._session and not self._session.closed:
            await self._session.close()
    
    async def login(self, email: str, password: str) -> Dict[str, Any]:
        """Авторизация пользователя"""
        session = await self.get_session()
        async with session.post(
            f"{self.auth_url}/login",
            json={"email": email, "password": password}
        ) as response:
            if response.status != 200:
                error = await response.json()
                raise Exception(error.get("error", "Ошибка авторизации"))
            return await response.json()
    
    async def get_user_info(self, token: str) -> Dict[str, Any]:
        """Получить информацию о пользователе"""
        session = await self.get_session()
        async with session.get(
            f"{self.auth_url}/me",
            headers={"Authorization": f"Bearer {token}"}
        ) as response:
            if response.status != 200:
                raise Exception("Не удалось получить информацию о пользователе")
            return await response.json()
    
    async def get_user_workflows(self, token: str) -> list:
        """Получить workflows пользователя"""
        session = await self.get_session()
        async with session.get(
            f"{self.auth_url}/workflows/my",
            headers={"Authorization": f"Bearer {token}"}
        ) as response:
            if response.status != 200:
                raise Exception("Не удалось получить workflows")
            data = await response.json()
            return data.get("data", [])
    
    async def create_task_and_stream(
        self,
        token: str,
        title: str,
        description: str,
        provider: str,
        model: str,
        api_key: str,
        workflow: Optional[Dict] = None,
        progress_callback=None
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """
        Создать задачу через WebSocket и получать обновления в реальном времени
        """
        # Формируем URL для WebSocket подключения
        # gateway_url это базовый URL (ws://apigateway:3111), нужно добавить путь
        ws_url = f"{self.gateway_url}/task/create"

        # Подготавливаем данные для задачи
        task_data = {
            "username": "user",
            "user_id": "",
            "title": title,
            "description": description,
            "tokens": {
                provider: api_key
            },
            "meta": {
                "provider": provider,
                "model": model
            }
        }
        
        # Добавляем workflow если есть
        if workflow:
            task_data["workflow"] = {
                "use_ai_planning": True,
                "architecture": workflow.get("category", ""),
                "tech_stack": workflow.get("tags", []),
                "managers": []
            }
        
        max_retries = 3
        for attempt in range(max_retries):
            try:
                print(f"[WebSocket] Попытка подключения #{attempt + 1} к {ws_url}")
                
                # Подключаемся по WebSocket
                async with websockets.connect(
                    ws_url,
                    ping_interval=20,
                    ping_timeout=10,
                    close_timeout=10
                ) as ws:
                    print("[WebSocket] Подключено, отправляю задачу...")
                    
                    # Отправляем данные задачи
                    await ws.send(json.dumps(task_data))
                    print("[WebSocket] Данные задачи отправлены")
                    
                    # Читаем стрим обновлений с таймаутом
                    while True:
                        try:
                            message = await asyncio.wait_for(ws.recv(), timeout=300)
                            update = json.loads(message)
                            
                            if progress_callback:
                                await progress_callback(update)
                            
                            yield update
                            
                            # Если задача завершена - выходим
                            if update.get("type") in ["success", "error"]:
                                print(f"[WebSocket] Задача завершена: {update.get('type')}")
                                return
                                
                        except asyncio.TimeoutError:
                            print("[WebSocket] Таймаут ожидания сообщения")
                            if progress_callback:
                                await progress_callback({
                                    "type": "error",
                                    "message": "Таймаут соединения"
                                })
                            yield {"type": "error", "message": "Таймаут соединения"}
                            return
                        except json.JSONDecodeError:
                            continue
                            
            except websockets.exceptions.ConnectionClosed as e:
                print(f"[WebSocket] Соединение закрыто: {e}")
                if attempt < max_retries - 1:
                    await asyncio.sleep(2)
                    continue
                if progress_callback:
                    await progress_callback({
                        "type": "error",
                        "message": f"Соединение закрыто: {str(e)}"
                    })
                yield {"type": "error", "message": f"Соединение закрыто: {str(e)}"}
            except Exception as e:
                print(f"[WebSocket] Ошибка: {e}")
                if attempt < max_retries - 1:
                    await asyncio.sleep(2)
                    continue
                if progress_callback:
                    await progress_callback({
                        "type": "error",
                        "message": f"Ошибка подключения: {str(e)}"
                    })
                yield {"type": "error", "message": f"Ошибка подключения: {str(e)}"}

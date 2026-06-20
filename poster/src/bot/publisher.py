from aiogram import Bot
from aiogram.types import User


class Publisher:
    def __init__(self, token: str, channel_id: str):
        self._bot = Bot(token=token)
        self._channel_id = channel_id

    async def post(self, text: str):
        await self._bot.send_message(
            self._channel_id,
            text,
            disable_web_page_preview=True,
        )

    async def get_me(self) -> User:
        return await self._bot.get_me()

    async def close(self):
        await self._bot.session.close()

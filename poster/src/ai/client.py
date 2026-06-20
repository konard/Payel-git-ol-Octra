import logging

from openai import AsyncOpenAI

logger = logging.getLogger(__name__)

_PLACEHOLDER = "sk-placeholder"


class AiClient:
    def __init__(self, endpoint: str, api_key: str, model: str):
        self._api_key = api_key
        self._model = model
        if not api_key:
            api_key = _PLACEHOLDER
            logger.warning("No API_KEY set — AI client will skip generation")
        self._client = AsyncOpenAI(
            base_url=endpoint,
            api_key=api_key,
        )

    async def generate(self, prompt: str) -> str | None:
        if not self._api_key:
            logger.info("AI skipped (no API key)")
            return None
        response = await self._client.chat.completions.create(
            model=self._model,
            messages=[{"role": "user", "content": prompt}],
        )
        return response.choices[0].message.content.strip()

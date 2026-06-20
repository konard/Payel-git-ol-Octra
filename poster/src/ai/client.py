from openai import AsyncOpenAI


class AiClient:
    def __init__(self, endpoint: str, api_key: str, model: str):
        self._client = AsyncOpenAI(
            base_url=endpoint,
            api_key=api_key,
        )
        self._model = model

    async def generate(self, prompt: str) -> str:
        response = await self._client.chat.completions.create(
            model=self._model,
            messages=[{"role": "user", "content": prompt}],
        )
        return response.choices[0].message.content.strip()

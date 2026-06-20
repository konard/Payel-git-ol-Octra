from dataclasses import dataclass
from pathlib import Path
from os import getenv

from dotenv import load_dotenv


@dataclass
class Config:
    tg_bot_token: str
    channel_id: str = "@octra_ai"
    ai_endpoint: str = "https://api.gonkagate.com/v1/chat/completions"
    ai_api_key: str = ""
    ai_model: str = "moonshotai/kimi-k2.6"
    poll_interval: int = 5400

    pg_host: str = "localhost"
    pg_port: int = 5432
    pg_user: str = "crewai"
    pg_password: str = "crewai_password"
    pg_db: str = "crewai"

    repo_root: Path = Path(__file__).resolve().parent.parent.parent
    state_dir: Path = Path(__file__).resolve().parent.parent

    def __post_init__(self):
        env_root = getenv("REPO_ROOT")
        if env_root:
            self.repo_root = Path(env_root)
        self.poster_file = self.repo_root / ".poster"
        self.ai_endpoint = self.ai_endpoint.replace("/chat/completions", "")

    @classmethod
    def from_env(cls) -> "Config":
        env_path = Path(__file__).resolve().parent.parent / ".env"
        load_dotenv(env_path)

        token = getenv("TG_BOT_TOKEN")
        if not token:
            raise ValueError("TG_BOT_TOKEN not set in .env")

        return cls(
            tg_bot_token=token,
            channel_id=getenv("CHANNEL_ID", "@octra_ai"),
            ai_endpoint=getenv("BASE_ULR", "https://api.gonkagate.com/v1/chat/completions"),
            ai_api_key=getenv("API_KEY", ""),
            ai_model=getenv("AI_MODEL", "moonshotai/kimi-k2.6"),
            poll_interval=int(getenv("POLL_INTERVAL", "5400")),
            pg_host=getenv("POSTGRES_HOST", "localhost"),
            pg_port=int(getenv("POSTGRES_PORT", "5432")),
            pg_user=getenv("POSTGRES_USER", "crewai"),
            pg_password=getenv("POSTGRES_PASSWORD", "crewai_password"),
            pg_db=getenv("POSTGRES_DB", "crewai"),
        )

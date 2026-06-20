import asyncpg
from src.git_observer import CommitInfo, CommitStatus


class Database:
    def __init__(self, dsn: str):
        self._dsn = dsn
        self._pool: asyncpg.Pool | None = None

    async def connect(self):
        self._pool = await asyncpg.create_pool(self._dsn, min_size=1, max_size=2)
        async with self._pool.acquire() as conn:
            await conn.execute("""
                CREATE TABLE IF NOT EXISTS posters (
                    commit_hash TEXT PRIMARY KEY,
                    branch TEXT NOT NULL DEFAULT '',
                    author TEXT NOT NULL DEFAULT '',
                    message TEXT NOT NULL DEFAULT '',
                    status TEXT NOT NULL DEFAULT 'local',
                    published BOOLEAN NOT NULL DEFAULT FALSE,
                    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                    published_at TIMESTAMPTZ
                )
            """)

    async def close(self):
        if self._pool:
            await self._pool.close()

    async def known_hashes(self) -> set[str]:
        async with self._pool.acquire() as conn:
            rows = await conn.fetch("SELECT commit_hash FROM posters")
            return {r["commit_hash"] for r in rows}

    async def save_commit(self, commit: CommitInfo, status: CommitStatus):
        async with self._pool.acquire() as conn:
            await conn.execute(
                """
                INSERT INTO posters (commit_hash, author, message, status)
                VALUES ($1, $2, $3, $4)
                ON CONFLICT (commit_hash) DO NOTHING
                """,
                commit.hash, commit.author, commit.message, status.value,
            )

    async def mark_published(self, hashes: list[str]):
        async with self._pool.acquire() as conn:
            await conn.execute(
                """
                UPDATE posters
                SET published = TRUE, published_at = NOW()
                WHERE commit_hash = ANY($1::text[])
                """,
                hashes,
            )

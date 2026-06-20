import asyncio
import os
import logging

import hashlib

from src.config import Config
from src.database import Database
from src.git_observer import CommitInfo, CommitStatus, GitObserver
from src.poster_watcher import PosterWatcher
from src.ai.client import AiClient
from src.ai import prompts
from src.bot.publisher import Publisher

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class PosterBot:
    def __init__(self):
        self.cfg = Config.from_env()
        self.db = Database(
            f"postgresql://{self.cfg.pg_user}:{self.cfg.pg_password}"
            f"@{self.cfg.pg_host}:{self.cfg.pg_port}/{self.cfg.pg_db}"
        )
        self.git = GitObserver(self.cfg.repo_root)
        self.poster = PosterWatcher(self.cfg.poster_file)
        self.ai = AiClient(
            self.cfg.ai_endpoint,
            self.cfg.ai_api_key,
            self.cfg.ai_model,
        )
        self.publisher = Publisher(self.cfg.tg_bot_token, self.cfg.channel_id)

    async def _scan_and_publish(self) -> bool:
        all_commits = self.git.all_commits()
        if not all_commits:
            return False

        known = await self.db.known_hashes()
        new_commits = [c for c in all_commits if c.hash not in known]
        if not new_commits:
            return False

        logger.info("%d new commit(s) found", len(new_commits))

        statuses = {}
        for c in new_commits:
            status = self.git.determine_status(c.hash)
            statuses[c.hash] = status
            await self.db.save_commit(c, status)

        prompt = prompts.for_new_commits(new_commits, statuses)
        text = await self.ai.generate(prompt)
        await self.publisher.post(text)
        logger.info("posted to %s", self.cfg.channel_id)

        await self.db.mark_published([c.hash for c in new_commits])
        return True

    async def _handle_poster(self) -> bool:
        if not self.poster.exists:
            return False

        content = self.poster.read_content()
        if not content:
            return False

        key = "poster:" + hashlib.md5(content.encode()).hexdigest()
        known = await self.db.known_hashes()
        if key in known:
            return False

        logger.info("new .poster content detected")

        prompt = prompts.for_poster(content)
        text = await self.ai.generate(prompt)
        await self.publisher.post(text)
        logger.info("posted .poster content")

        await self.db.save_commit(
            CommitInfo(key, "poster", "poster"),
            CommitStatus.DONE,
        )
        return True

    async def run(self):
        await self.db.connect()
        me = await self.publisher.get_me()
        logger.info("bot @%s → %s", me.username, self.cfg.channel_id)

        while True:
            try:
                posted = await self._scan_and_publish()
                if not posted:
                    await self._handle_poster()
            except Exception as e:
                logger.error("check failed: %s", e, exc_info=True)

            await asyncio.sleep(
                5 if posted else self.cfg.poll_interval
            )

    async def shutdown(self):
        await self.publisher.close()
        await self.db.close()


async def main():
    bot = PosterBot()

    if os.name != "nt":
        import signal
        loop = asyncio.get_event_loop()
        for sig in (signal.SIGINT, signal.SIGTERM):
            loop.add_signal_handler(sig, lambda: asyncio.create_task(bot.shutdown()))

    try:
        await bot.run()
    except (KeyboardInterrupt, asyncio.CancelledError):
        logger.info("stopped")
    finally:
        await bot.shutdown()


if __name__ == "__main__":
    asyncio.run(main())

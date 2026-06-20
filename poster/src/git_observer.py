import subprocess
from dataclasses import dataclass
from pathlib import Path
from enum import Enum


class CommitStatus(str, Enum):
    LOCAL = "local"
    DEV = "dev"
    DONE = "done"


@dataclass(frozen=True)
class CommitInfo:
    hash: str
    author: str
    message: str


class GitObserver:
    def __init__(self, repo_root: Path):
        self._root = repo_root

    def _run(self, *args: str) -> str:
        result = subprocess.run(
            ["git", *args],
            capture_output=True, text=True,
            cwd=str(self._root),
        )
        return result.stdout.strip()

    @property
    def current_head(self) -> str:
        return self._run("rev-parse", "HEAD")

    @property
    def current_branch(self) -> str:
        return self._run("rev-parse", "--abbrev-ref", "HEAD")

    def new_commits_since(self, commit_hash: str | None) -> list[CommitInfo]:
        if not commit_hash:
            return []

        log = self._run(
            "log", f"{commit_hash}..HEAD",
            "--format=%H|||%an|||%s",
            "--reverse",
        )
        if not log:
            return []

        result: list[CommitInfo] = []
        for line in log.split("\n"):
            line = line.strip()
            if not line:
                continue
            parts = line.split("|||", 2)
            if len(parts) == 3:
                result.append(CommitInfo(parts[0], parts[1], parts[2]))

        return result

    def all_commits(self) -> list[CommitInfo]:
        log = self._run(
            "log", "--all",
            "--format=%H|||%an|||%s",
            "--reverse",
        )
        if not log:
            return []

        result: list[CommitInfo] = []
        for line in log.split("\n"):
            line = line.strip()
            if not line:
                continue
            parts = line.split("|||", 2)
            if len(parts) == 3:
                result.append(CommitInfo(parts[0], parts[1], parts[2]))

        return result

    def determine_status(self, commit_hash: str) -> CommitStatus:
        branches = self._run("branch", "-a", "--contains", commit_hash)
        if not branches:
            return CommitStatus.LOCAL

        for line in branches.split("\n"):
            line = line.strip().lstrip("* ")
            if line in ("master", "remotes/origin/master"):
                return CommitStatus.DONE

        for line in branches.split("\n"):
            line = line.strip().lstrip("* ")
            if line.startswith("remotes/origin/"):
                return CommitStatus.DEV

        return CommitStatus.LOCAL

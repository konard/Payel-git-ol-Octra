import json
import hashlib
from pathlib import Path


class StateManager:
    def __init__(self, state_dir: Path):
        self._file = state_dir / ".state.json"
        self._data: dict = self._load()

    def _load(self) -> dict:
        if self._file.exists():
            return json.loads(self._file.read_text())
        return {}

    def _save(self):
        self._file.write_text(json.dumps(self._data, indent=2))

    def last_commit(self, branch: str) -> str | None:
        return self._data.get(f"commit:{branch}")

    def save_last_commit(self, branch: str, hash_: str):
        self._data[f"commit:{branch}"] = hash_
        self._save()

    @property
    def poster_hash(self) -> str | None:
        return self._data.get("poster_hash")

    @poster_hash.setter
    def poster_hash(self, value: str):
        self._data["poster_hash"] = value
        self._save()

    @staticmethod
    def compute_hash(content: str) -> str:
        return hashlib.md5(content.encode()).hexdigest()

from pathlib import Path


class PosterWatcher:
    def __init__(self, file_path: Path):
        self._path = file_path

    @property
    def exists(self) -> bool:
        return self._path.exists()

    def read_content(self) -> str:
        return self._path.read_text(encoding="utf-8").strip()

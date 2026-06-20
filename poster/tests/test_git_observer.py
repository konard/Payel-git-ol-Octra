from pathlib import Path
from src.git_observer import CommitInfo, CommitStatus, GitObserver


def test_parse_commit_info():
    info = CommitInfo("abc123", "Alice", "fix: resolve bug")
    assert info.hash == "abc123"
    assert info.author == "Alice"
    assert info.message == "fix: resolve bug"


def test_all_commits_parses_git_log(monkeypatch):
    fake_log = (
        "abc123|||Alice|||feat: add login\n"
        "def456|||Bob|||fix: typo\n"
    )

    def mock_run(*args, **kwargs):
        return fake_log

    observer = GitObserver(Path("/tmp"))
    monkeypatch.setattr(observer, "_run", mock_run)

    commits = observer.all_commits()
    assert len(commits) == 2
    assert commits[0].hash == "abc123"
    assert commits[1].message == "fix: typo"


def test_determine_status_local(monkeypatch):
    def mock_run(*args, **kwargs):
        return "* feature-x"

    observer = GitObserver(Path("/tmp"))
    monkeypatch.setattr(observer, "_run", mock_run)

    status = observer.determine_status("abc123")
    assert status == CommitStatus.LOCAL


def test_determine_status_dev(monkeypatch):
    def mock_run(*args, **kwargs):
        return "  remotes/origin/feature-x\n  feature-x"

    observer = GitObserver(Path("/tmp"))
    monkeypatch.setattr(observer, "_run", mock_run)

    status = observer.determine_status("abc123")
    assert status == CommitStatus.DEV


def test_determine_status_done_via_origin_master(monkeypatch):
    def mock_run(*args, **kwargs):
        return "  remotes/origin/master\n  remotes/origin/feature-x"

    observer = GitObserver(Path("/tmp"))
    monkeypatch.setattr(observer, "_run", mock_run)

    status = observer.determine_status("abc123")
    assert status == CommitStatus.DONE


def test_determine_status_done_via_local_master(monkeypatch):
    def mock_run(*args, **kwargs):
        return "* master\n  remotes/origin/feature-x"

    observer = GitObserver(Path("/tmp"))
    monkeypatch.setattr(observer, "_run", mock_run)

    status = observer.determine_status("abc123")
    assert status == CommitStatus.DONE

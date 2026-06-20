from src.git_observer import CommitInfo, CommitStatus
from src.ai.prompts import for_new_commits


def test_for_new_commits_contains_last_commit():
    commits = [
        CommitInfo("a", "Alice", "feat: initial"),
        CommitInfo("b", "Bob", "fix: bug"),
    ]
    statuses = {"a": CommitStatus.LOCAL, "b": CommitStatus.DEV}
    prompt = for_new_commits(commits, statuses)

    assert "fix: bug" in prompt
    assert "Последнее изменение" in prompt


def test_for_new_commits_contains_status_labels():
    commits = [
        CommitInfo("a", "Alice", "feat: x"),
    ]
    statuses = {"a": CommitStatus.DONE}
    prompt = for_new_commits(commits, statuses)

    assert "Готово" in prompt


def test_for_new_commits_shows_messages_without_author():
    commits = [
        CommitInfo("a", "Payel-git-ol", "feat: x"),
        CommitInfo("b", "konard", "fix: y"),
    ]
    statuses = {"a": CommitStatus.LOCAL, "b": CommitStatus.DEV}
    prompt = for_new_commits(commits, statuses)

    assert "feat: x" in prompt
    assert "fix: y" in prompt


def test_for_new_commits_has_no_mention_instruction():
    commits = [
        CommitInfo("a", "Payel-git-ol", "feat: x"),
    ]
    statuses = {"a": CommitStatus.LOCAL}
    prompt = for_new_commits(commits, statuses)

    assert "Не упоминай" in prompt
    assert "Payel-git-ol" in prompt
    assert "konard" in prompt

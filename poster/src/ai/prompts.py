from src.git_observer import CommitInfo, CommitStatus


def for_new_commits(commits: list[CommitInfo], statuses: dict[str, CommitStatus]) -> str:
    status_labels = {
        CommitStatus.LOCAL: "👨‍💻 Локальная разработка",
        CommitStatus.DEV: "🔧 В разработке",
        CommitStatus.DONE: "✅ Готово",
    }

    last = commits[-1]

    groups: dict[str, list[str]] = {}
    for c in commits:
        label = status_labels.get(statuses.get(c.hash, CommitStatus.LOCAL))
        groups.setdefault(label, []).append(f"• {c.message}")

    parts = []
    for label, items in groups.items():
        parts.append(f"{label}\n" + "\n".join(items))

    return (
        "Ты — копирайтер Telegram-канала проекта Octra. "
        "Octra — AI-оркестратор для сборки проектов на Nix.\n\n"
        f"Последнее изменение: {last.message}\n\n"
        "Все свежие изменения в репозитории:\n\n"
        + "\n\n".join(parts) +
        "\n\nНапиши пост. Главная новость — последнее изменение выше. "
        "Если это master → пиши как о релизе (📦 В Octra появилось). "
        "Если ветка → как о разработке (🔧 Работаем над). "
        "2–4 абзаца, с эмодзи. Без markdown-разметки. "
        "Не упоминай пользователей Payel-git-ol и konard. "
        "В конце: 💬 Чат: https://t.me/octra_ai"
    )


def for_poster(content: str) -> str:
    return (
        "Ты — копирайтор Telegram-канала проекта Octra. "
        "Octra — AI-оркестратор для сборки проектов на Nix.\n\n"
        "Напиши пост на основе текста ниже. "
        "Стиль: длинные информационные посты, как в примере ниже. "
        "С заголовком, несколькими абзацами, эмодзи, ссылками, "
        "разделителями и контактами в конце.\n\n"
        "Пример стиля:\n"
        "Заголовок\n\n"
        "Абзац с новостью.\n\n"
        "Ещё абзац с деталями и ссылкой — https://example.com\n\n"
        "________________________________\n"
        "💬 Чат проекта: https://t.me/chat\n"
        "📹 YouTube: https://youtube.com/@channel\n\n\n"
        "Текст для поста:\n"
        f"{content}\n\n"
        "Напиши пост в таком стиле. Без markdown-разметки (кроме ссылок). "
        "Используй эмодзи. Если в тексте есть даты/события — оформи списком с ➖."
    )

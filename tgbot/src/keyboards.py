"""
Клавиатуры для бота
"""
from aiogram.types import InlineKeyboardMarkup, InlineKeyboardButton, ReplyKeyboardMarkup, KeyboardButton
from aiogram.utils.keyboard import InlineKeyboardBuilder, ReplyKeyboardBuilder


def get_start_keyboard() -> InlineKeyboardMarkup:
    """Клавиатура для стартового сообщения"""
    builder = InlineKeyboardBuilder()
    builder.button(text="🔐 Войти в аккаунт", callback_data="auth:login")
    return builder.as_markup()


def get_login_keyboard() -> InlineKeyboardMarkup:
    """Клавиатура для входа"""
    builder = InlineKeyboardBuilder()
    builder.button(text="👤 У меня есть аккаунт", callback_data="auth:login:email")
    return builder.as_markup()


def get_workflow_keyboard(workflows: list) -> InlineKeyboardMarkup:
    """Клавиатура с workflows пользователя"""
    builder = InlineKeyboardBuilder()
    for wf in workflows:
        builder.button(
            text=f"📋 {wf.get('name', 'Без названия')}",
            callback_data=f"workflow:select:{wf['id']}"
        )
    builder.adjust(1)
    return builder.as_markup()


def get_provider_keyboard() -> InlineKeyboardMarkup:
    """Клавиатура с выбором провайдера"""
    builder = InlineKeyboardBuilder()
    providers = [
        ("🟠 OpenRouter", "openrouter"),
        ("🔵 Google Gemini", "gemini"),
        ("🟢 OpenAI", "openai"),
        ("🟣 Anthropic Claude", "claude"),
        ("🔵 DeepSeek", "deepseek"),
        ("🟡 xAI Grok", "grok"),
        ("🟣 Qwen", "qwen"),
    ]
    for text, callback in providers:
        builder.button(text=text, callback_data=f"provider:select:{callback}")
    builder.adjust(1)
    return builder.as_markup()


def get_model_keyboard(provider: str, page: int = 0) -> InlineKeyboardMarkup:
    """Клавиатура с моделями для выбранного провайдера"""
    builder = InlineKeyboardBuilder()
    
    models = {
        "openrouter": [
            # OpenAI GPT-5.x
            ("GPT-5.4", "openai/gpt-5.4"),
            ("GPT-5.4 Pro", "openai/gpt-5.4-pro"),
            ("GPT-5.4 Mini", "openai/gpt-5.4-mini"),
            ("GPT-5.4 Nano", "openai/gpt-5.4-nano"),
            ("GPT-5.4 Codex", "openai/gpt-5.4-codex"),
            ("GPT-5.3 Codex", "openai/gpt-5.3-codex"),
            ("GPT-5.2", "openai/gpt-5.2"),
            ("GPT-5.2 Codex", "openai/gpt-5.2-codex"),
            ("GPT-5.1 Medium", "openai/gpt-5.1"),
            ("GPT-5.1 High", "openai/gpt-5.1-high"),
            ("GPT-5.1 Codex", "openai/gpt-5.1-codex"),
            ("GPT-5.1 Instant", "openai/gpt-5.1-instant"),
            ("GPT-5.1 Thinking", "openai/gpt-5.1-thinking"),
            ("GPT-5", "openai/gpt-5"),
            ("GPT-5 Mini", "openai/gpt-5-mini"),
            ("GPT-5 Nano", "openai/gpt-5-nano"),
            ("GPT-5 High", "openai/gpt-5-high"),
            ("GPT-5 Medium", "openai/gpt-5-medium"),
            ("GPT-5 Codex", "openai/gpt-5-codex"),
            # GPT-4.x
            ("GPT-4.1", "openai/gpt-4.1"),
            ("GPT-4.1 Mini", "openai/gpt-4.1-mini"),
            ("GPT-4.1 Nano", "openai/gpt-4.1-nano"),
            ("GPT-4.5", "openai/gpt-4.5-preview"),
            ("GPT-4o", "openai/gpt-4o"),
            ("GPT-4o Mini", "openai/gpt-4o-mini"),
            ("GPT-4 Turbo", "openai/gpt-4-turbo"),
            ("GPT-4", "openai/gpt-4"),
            ("GPT-3.5 Turbo", "openai/gpt-3.5-turbo"),
            # Claude
            ("Claude Opus 4.6", "anthropic/claude-opus-4-6"),
            ("Claude Opus 4.5", "anthropic/claude-opus-4-5"),
            ("Claude Sonnet 4.6", "anthropic/claude-sonnet-4-6"),
            ("Claude Sonnet 4.5", "anthropic/claude-sonnet-4-5"),
            ("Claude Sonnet 4", "anthropic/claude-sonnet-4-20250514"),
            ("Claude Haiku 4.5", "anthropic/claude-haiku-4-5"),
            # Gemini
            ("Gemini 3.1 Pro", "google/gemini-3.1-pro"),
            ("Gemini 3 Pro", "google/gemini-3-pro"),
            ("Gemini 3 Flash", "google/gemini-3-flash"),
            ("Gemini 2.5 Pro", "google/gemini-2.5-pro"),
            ("Gemini 2.5 Flash", "google/gemini-2.5-flash"),
            ("Gemini 2.5 Flash-Lite", "google/gemini-2.5-flash-lite"),
            ("Gemini 2.0 Flash", "google/gemini-2.0-flash"),
            ("Gemini 2.0 Flash-Lite", "google/gemini-2.0-flash-lite"),
            ("Gemini 1.5 Pro", "google/gemini-1.5-pro"),
            ("Gemini 1.5 Flash", "google/gemini-1.5-flash"),
            ("Gemini 1.0 Pro", "google/gemini-1.0-pro"),
            # DeepSeek
            ("DeepSeek V3.2 Thinking", "deepseek/deepseek-v3.2"),
            ("DeepSeek V3.2", "deepseek/deepseek-v3.2-nonthinking"),
            ("DeepSeek V3.2 Speciale", "deepseek/deepseek-v3.2-speciale"),
            ("DeepSeek V3.2 Exp", "deepseek/deepseek-v3.2-exp"),
            ("DeepSeek V3.1", "deepseek/deepseek-v3.1"),
            ("DeepSeek V3", "deepseek/deepseek-v3"),
            ("DeepSeek R1", "deepseek/deepseek-r1"),
            ("DeepSeek R1 0528", "deepseek/deepseek-r1-0528"),
            ("DeepSeek R1 Zero", "deepseek/deepseek-r1-zero"),
            ("DeepSeek R1 Distill Llama 70B", "deepseek/deepseek-r1-distill-llama-70b"),
            ("DeepSeek R1 Distill Qwen 32B", "deepseek/deepseek-r1-distill-qwen-32b"),
            ("DeepSeek R1 Distill Qwen 14B", "deepseek/deepseek-r1-distill-qwen-14b"),
            ("DeepSeek R1 Distill Qwen 7B", "deepseek/deepseek-r1-distill-qwen-7b"),
            ("DeepSeek R1 Distill Qwen 1.5B", "deepseek/deepseek-r1-distill-qwen-1.5b"),
            ("DeepSeek VL2", "deepseek/deepseek-vl2"),
            ("DeepSeek VL2 Small", "deepseek/deepseek-vl2-small"),
            ("DeepSeek VL2 Tiny", "deepseek/deepseek-vl2-tiny"),
            # Grok
            ("Grok 4.20", "x-ai/grok-4.20"),
            ("Grok 4.1 Fast Reasoning", "x-ai/grok-4.1"),
            ("Grok 4.1 Non-Reasoning", "x-ai/grok-4.1-nonreasoning"),
            ("Grok 4 Heavy", "x-ai/grok-4-heavy"),
            ("Grok 4 Fast Reasoning", "x-ai/grok-4-fast"),
            ("Grok 4 Fast Non-Reasoning", "x-ai/grok-4-fast-nonreasoning"),
            ("Grok Code Fast 1", "x-ai/grok-code-fast-1"),
            ("Grok 3", "x-ai/grok-3"),
            ("Grok 3 Mini", "x-ai/grok-3-mini"),
            ("Grok 2", "x-ai/grok-2"),
            ("Grok 2 Mini", "x-ai/grok-2-mini"),
            ("Grok 1.5V", "x-ai/grok-1.5v"),
            ("Grok 1.5", "x-ai/grok-1.5"),
            # Qwen
            ("Qwen3 Max", "qwen/qwen3-max"),
            ("Qwen3 VL 32B Thinking", "qwen/qwen3-vl-32b-thinking"),
            ("Qwen3 Next 80B A3B Instruct", "qwen/qwen3-next-80b-a3b-instruct"),
            ("Qwen3 235B A22B Thinking", "qwen/qwen3-235b-a22b-thinking"),
            ("Qwen3 235B A22B Instruct", "qwen/qwen3-235b-a22b-instruct"),
            ("Qwen3 32B", "qwen/qwen3-32b"),
            ("Qwen3 30B A3B", "qwen/qwen3-30b-a3b"),
            ("Qwen3 Coder ⭐", "qwen/qwen3-coder"),
            ("Qwen3 Coder 480B A35B", "qwen/qwen3-coder-480b-a35b-instruct"),
            ("Qwen2.5 Omni 7B", "qwen/qwen2.5-omni-7b"),
            ("Qwen2.5 VL 72B", "qwen/qwen2.5-vl-72b-instruct"),
            ("Qwen2.5 VL 32B", "qwen/qwen2.5-vl-32b-instruct"),
            ("Qwen2.5 VL 7B", "qwen/qwen2.5-vl-7b-instruct"),
            ("Qwen2 VL 72B Instruct", "qwen/qwen2-vl-72b-instruct"),
            ("QwQ 32B", "qwen/qwq-32b"),
            ("QvQ 72B Preview", "qwen/qvq-72b-preview"),
            # Free models
            ("Gemini Exp 1206 [FREE]", "google/gemini-exp-1206:free"),
            ("Gemini 2.0 Flash Exp [FREE]", "google/gemini-2.0-flash-exp:free"),
            ("Qwen 3.6 Plus [FREE]", "qwen/qwen3.6-plus:free"),
        ],
        "gemini": [
            ("Gemini 2.5 Flash ⭐", "gemini-2.5-flash"),
            ("Gemini 2.5 Pro", "gemini-2.5-pro"),
            ("Gemini 2.0 Flash", "gemini-2.0-flash"),
            ("Gemini 2.0 Flash Lite", "gemini-2.0-flash-lite"),
            ("Gemini 2.0 Thinking", "gemini-2.0-flash-thinking-exp"),
            ("Gemini 1.5 Flash", "gemini-1.5-flash"),
            ("Gemini 1.5 Pro", "gemini-1.5-pro"),
        ],
        "openai": [
            ("GPT-4o Mini ⭐", "gpt-4o-mini"),
            ("GPT-4o", "gpt-4o"),
            ("GPT-4 Turbo", "gpt-4-turbo"),
            ("o1 Mini", "o1-mini"),
            ("o1", "o1"),
            ("o3 Mini", "o3-mini"),
            ("o4 Mini", "o4-mini"),
        ],
        "claude": [
            ("Claude Opus 4.6", "anthropic/claude-opus-4-6"),
            ("Claude Opus 4.5", "anthropic/claude-opus-4-5"),
            ("Claude Sonnet 4.6", "anthropic/claude-sonnet-4-6"),
            ("Claude Sonnet 4.5", "anthropic/claude-sonnet-4-5"),
            ("Claude Sonnet 4", "anthropic/claude-sonnet-4-20250514"),
            ("Claude Haiku 4.5", "anthropic/claude-haiku-4-5"),
        ],
        "deepseek": [
            ("DeepSeek Chat V3 ⭐", "deepseek-chat"),
            ("DeepSeek R1", "deepseek-reasoner"),
            ("DeepSeek V2.5", "deepseek-v2.5"),
            ("DeepSeek Coder", "deepseek-coder"),
        ],
        "grok": [
            ("Grok 3 ⭐", "grok-3"),
            ("Grok 2", "grok-2"),
            ("Grok Beta", "grok-beta"),
        ],
        "qwen": [
            ("Qwen 3.6 Plus (Free) ⭐", "qwen/qwen3.6-plus:free"),
            ("Qwen3 Coder", "qwen/qwen3-coder"),
            ("Qwen3 Max", "qwen/qwen3-max"),
            ("Qwen3 VL 32B Thinking", "qwen/qwen3-vl-32b-thinking"),
            ("Qwen3 Next 80B A3B Instruct", "qwen/qwen3-next-80b-a3b-instruct"),
            ("Qwen3 235B A22B Thinking", "qwen/qwen3-235b-a22b-thinking"),
            ("Qwen3 235B A22B Instruct", "qwen/qwen3-235b-a22b-instruct"),
            ("Qwen3 32B", "qwen/qwen3-32b"),
            ("Qwen3 30B A3B", "qwen/qwen3-30b-a3b"),
            ("Qwen3 Coder 480B A35B Instruct", "qwen/qwen3-coder-480b-a35b-instruct"),
            ("Qwen2.5 Omni 7B", "qwen/qwen2.5-omni-7b"),
            ("Qwen2.5 VL 72B", "qwen/qwen2.5-vl-72b-instruct"),
            ("Qwen2.5 VL 32B", "qwen/qwen2.5-vl-32b-instruct"),
            ("Qwen2.5 VL 7B", "qwen/qwen2.5-vl-7b-instruct"),
            ("Qwen2 VL 72B Instruct", "qwen/qwen2-vl-72b-instruct"),
            ("QwQ 32B", "qwen/qwq-32b"),
            ("QvQ 72B Preview", "qwen/qvq-72b-preview"),
        ],
    }
    
    provider_models = models.get(provider, [])
    
    # Пагинация: максимум 70 символов на callback, максимум ~15 кнопок на страницу
    models_per_page = 15
    total_pages = (len(provider_models) + models_per_page - 1) // models_per_page
    
    start_idx = page * models_per_page
    end_idx = min(start_idx + models_per_page, len(provider_models))
    
    for text, model_id in provider_models[start_idx:end_idx]:
        # Обрезаем длинные названия
        display_text = text[:50] + "..." if len(text) > 50 else text
        builder.button(text=display_text, callback_data=f"model:select:{model_id}")
    
    # Добавляем навигацию по страницам если нужно
    if total_pages > 1:
        nav_buttons = InlineKeyboardBuilder()
        if page > 0:
            nav_buttons.button(text="⬅️ Назад", callback_data=f"model:page:{provider}:{page-1}")
        if page < total_pages - 1:
            nav_buttons.button(text="Вперёд ➡️", callback_data=f"model:page:{provider}:{page+1}")
        nav_buttons.adjust(2)
        
        # Добавляем текст с номером страницы
        builder.button(text=f"📄 Страница {page + 1}/{total_pages}", callback_data="noop")
        builder.attach(nav_buttons)
    
    builder.adjust(1)
    return builder.as_markup()


def get_back_keyboard(action: str) -> InlineKeyboardMarkup:
    """Клавиатура с кнопкой назад"""
    builder = InlineKeyboardBuilder()
    builder.button(text="⬅️ Назад", callback_data=f"back:{action}")
    return builder.as_markup()


def get_cancel_keyboard() -> InlineKeyboardMarkup:
    """Клавиатура отмены"""
    builder = InlineKeyboardBuilder()
    builder.button(text="❌ Отмена", callback_data="cancel")
    return builder.as_markup()

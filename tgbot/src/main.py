"""
Telegram Bot для CrewAI
Главный файл запуска
"""
import asyncio
import logging
from aiogram import Bot, Dispatcher, F
from aiogram.client.default import DefaultBotProperties
from aiogram.fsm.storage.memory import MemoryStorage
from aiogram.fsm.context import FSMContext
from aiogram.types import BotCommand, Message

from src.config import BOT_TOKEN
from src.handlers.auth_handler import router as auth_router
from src.handlers.task_handler import router as task_router
from src.keyboards import get_start_keyboard
from src.handlers.task_handler import show_workflow_selection

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


async def set_commands(bot: Bot):
    """Установка команд бота"""
    commands = [
        BotCommand(command="start", description="🚀 Начать работу с ботом"),
        BotCommand(command="help", description="❓ Показать справку"),
        BotCommand(command="cancel", description="❌ Отменить текущую операцию"),
    ]
    await bot.set_my_commands(commands)


async def main():
    """Главная функция"""
    logger.info("Запуск CrewAI Telegram Bot...")

    # Создаём бота и диспетчер
    # Устанавливаем parse_mode="HTML" по умолчанию для всех сообщений
    bot = Bot(
        token=BOT_TOKEN,
        default=DefaultBotProperties(parse_mode="HTML")
    )
    storage = MemoryStorage()
    dp = Dispatcher(storage=storage)
    
    # Устанавливаем команды
    await set_commands(bot)
    
    # Подключаем роутеры
    dp.include_router(auth_router)
    dp.include_router(task_router)
    
    # Обработчик команды /start
    @dp.message(F.text == "/start")
    async def cmd_start(message: Message, state: FSMContext):
        """Обработка команды /start"""
        # Проверяем, есть ли данные авторизации
        data = await state.get_data()
        token = data.get("token")
        username = data.get("username")
        
        if token and username:
            # Пользователь уже авторизован - сразу показываем выбор workflow
            await message.answer(
                f"👋 <b>С возвращением, {username}!</b>\n\n"
                "Выберите workflow для новой задачи:",
            )
            await show_workflow_selection(message, state)
        else:
            # Нужно авторизоваться
            await message.answer(
                "👋 <b>Добро пожаловать в CrewAI Bot!</b>\n\n"
                "Этот бот позволяет:\n"
                "• Авторизоваться в вашем аккаунте\n"
                "• Выбрать workflow\n"
                "• Создать и выполнить задачу\n\n"
                "Нажмите кнопку ниже, чтобы войти в аккаунт:",
                reply_markup=get_start_keyboard()
            )
    
    # Обработчик команды /help
    @dp.message(F.text == "/help")
    async def cmd_help(message: Message):
        """Обработка команды /help"""
        await message.answer(
            "📖 <b>Справка по боту CrewAI</b>\n\n"
            "<b>Как использовать бота:</b>\n\n"
            "1️⃣ Нажмите 'Войти в аккаунт'\n"
            "2️⃣ Введите email и пароль\n"
            "3️⃣ Выберите workflow (или пропустите)\n"
            "4️⃣ Выберите провайдера и модель\n"
            "5️⃣ Введите API ключ\n"
            "6️⃣ Опишите задачу\n"
            "7️⃣ Дождитесь результата\n\n"
            "<b>Команды:</b>\n"
            "/start - Начать работу\n"
            "/help - Показать справку\n"
            "/cancel - Отменить операцию\n\n"
            "<b>Поддерживаемые провайдеры:</b>\n"
            "• OpenRouter (80+ моделей)\n"
            "• Google Gemini\n"
            "• OpenAI\n"
            "• Anthropic Claude\n"
            "• DeepSeek\n"
            "• xAI Grok\n"
            "• Qwen"
        )
    
    # Обработчик команды /cancel
    @dp.message(F.text == "/cancel")
    async def cmd_cancel(message: Message, state: FSMContext):
        """Обработка команды /cancel"""
        await state.clear()
        await message.answer(
            "❌ Операция отменена.\n\n"
            "Нажмите /start, чтобы начать заново."
        )
    
    # Запуск polling
    try:
        logger.info("Бот запущен и готов к работе!")
        await dp.start_polling(bot)
    except Exception as e:
        logger.error(f"Ошибка при запуске бота: {e}")
    finally:
        await bot.session.close()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("Бот остановлен пользователем")
    except Exception as e:
        logger.error(f"Критическая ошибка: {e}")

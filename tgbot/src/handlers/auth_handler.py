"""
Обработчики авторизации
"""
from aiogram import Router, F
from aiogram.types import Message, CallbackQuery
from aiogram.fsm.context import FSMContext
from src.states import AuthState
from src.keyboards import get_login_keyboard, get_back_keyboard
from src.services.api_service import APIService

router = Router()
api_service = APIService()


@router.callback_query(F.data == "auth:login")
async def process_auth_start(callback: CallbackQuery, state: FSMContext):
    """Начало процесса авторизации"""
    await callback.message.edit_text(
        "🔐 <b>Вход в аккаунт</b>\n\n"
        "Введите ваш email:",
        reply_markup=get_back_keyboard("start")
    )
    await state.set_state(AuthState.waiting_for_email)


@router.callback_query(F.data.startswith("back:"))
async def process_back(callback: CallbackQuery, state: FSMContext):
    """Кнопка назад"""
    action = callback.data.split(":")[1]
    
    if action == "start":
        await state.clear()
        await callback.message.edit_text(
            "👋 <b>Добро пожаловать в CrewAI Bot!</b>\n\n"
            "Этот бот позволяет:\n"
            "• Авторизоваться в вашем аккаунте\n"
            "• Выбрать workflow\n"
            "• Создать и выполнить задачу\n\n"
            "Нажмите кнопку ниже, чтобы войти в аккаунт:",
            reply_markup=get_login_keyboard()
        )


@router.message(AuthState.waiting_for_email)
async def process_email(message: Message, state: FSMContext):
    """Получение email"""
    email = message.text.strip()
    
    if "@" not in email:
        await message.answer("❌ Пожалуйста, введите корректный email:")
        return
    
    await state.update_data(email=email)
    await message.answer(
        "🔑 Теперь введите ваш пароль:",
        reply_markup=get_back_keyboard("email")
    )
    await state.set_state(AuthState.waiting_for_password)


@router.message(AuthState.waiting_for_password)
async def process_password(message: Message, state: FSMContext):
    """Получение пароля и авторизация"""
    password = message.text
    data = await state.get_data()
    email = data.get("email")
    
    # Показываем "загрузку"
    loading_msg = await message.answer("⏳ Выполняется вход...")
    
    try:
        # Авторизация
        auth_response = await api_service.login(email, password)
        
        # Сохраняем токен и информацию о пользователе
        token = auth_response["data"]["access_token"]
        user_info = await api_service.get_user_info(token)
        
        await state.update_data(
            token=token,
            user_id=user_info["data"]["user_id"],
            username=user_info["data"]["username"],
            email=user_info["data"]["email"]
        )
        
        await loading_msg.delete()
        
        await message.answer(
            f"✅ <b>Добро пожаловать, {user_info['data']['username']}!</b>\n\n"
            "Вы успешно вошли в аккаунт.\n"
            "Теперь давайте создадим задачу 🚀",
        )
        
        # Переходим к выбору workflow
        from src.handlers.task_handler import show_workflow_selection
        await show_workflow_selection(message, state)
        
    except Exception as e:
        await loading_msg.delete()
        await message.answer(
            f"❌ <b>Ошибка авторизации</b>\n\n"
            f"{str(e)}\n\n"
            "Попробуйте ещё раз:",
            reply_markup=get_back_keyboard("email")
        )
        return

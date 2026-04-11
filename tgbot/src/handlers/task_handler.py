"""
Обработчики создания задачи
"""
from aiogram import Router, F
from aiogram.types import Message, CallbackQuery
from aiogram.fsm.context import FSMContext
from src.states import TaskState, AuthState
from src.keyboards import (
    get_workflow_keyboard,
    get_provider_keyboard,
    get_model_keyboard,
    get_back_keyboard,
    get_cancel_keyboard
)
from src.services.api_service import APIService
import asyncio
import json

router = Router()
api_service = APIService()


async def show_workflow_selection(message, state: FSMContext):
    """Показать выбор workflow"""
    try:
        data = await state.get_data()
        token = data.get("token")
        
        workflows = await api_service.get_user_workflows(token)
        
        if not workflows:
            await message.answer(
                "⚠️ <b>У вас пока нет workflow</b>\n\n"
                "Создайте workflow через веб-интерфейс и попробуйте снова.\n\n"
                "Или мы можем создать задачу без workflow:"
            )
            await state.set_state(TaskState.waiting_for_title)
            return
        
        await message.answer(
            "📋 <b>Выберите workflow</b>\n\n"
            "Выберите один из ваших сохранённых workflow:",
            reply_markup=get_workflow_keyboard(workflows)
        )
        await state.set_state(TaskState.selecting_workflow)
        
    except Exception as e:
        await message.answer(f"❌ Ошибка: {str(e)}")


@router.callback_query(F.data.startswith("workflow:select:"))
async def process_workflow_selection(callback: CallbackQuery, state: FSMContext):
    """Выбор workflow"""
    workflow_id = callback.data.split(":")[2]
    
    data = await state.get_data()
    token = data.get("token")
    
    # Получаем workflows и находим выбранный
    workflows = await api_service.get_user_workflows(token)
    selected_wf = next((wf for wf in workflows if wf["id"] == workflow_id), None)
    
    if selected_wf:
        await state.update_data(
            workflow_id=workflow_id,
            workflow=selected_wf
        )
    
    await callback.message.edit_text(
        "⚙️ <b>Выберите AI провайдера</b>\n\n"
        "Какой провайдер вы хотите использовать?",
        reply_markup=get_provider_keyboard()
    )
    await state.set_state(TaskState.selecting_provider)


@router.callback_query(F.data.startswith("provider:select:"))
async def process_provider_selection(callback: CallbackQuery, state: FSMContext):
    """Выбор провайдера"""
    provider = callback.data.split(":")[2]
    
    await state.update_data(provider=provider)
    
    provider_names = {
        "openrouter": "OpenRouter",
        "gemini": "Google Gemini",
        "openai": "OpenAI",
        "claude": "Anthropic Claude",
        "deepseek": "DeepSeek",
        "grok": "xAI Grok",
        "qwen": "Qwen"
    }
    
    await callback.message.edit_text(
        f"🤖 <b>Выберите модель {provider_names.get(provider, provider)}</b>\n\n"
        "Какую модель вы хотите использовать?",
        reply_markup=get_model_keyboard(provider, page=0)
    )
    await state.set_state(TaskState.selecting_model)


@router.callback_query(F.data.startswith("model:page:"))
async def handle_model_pagination(callback: CallbackQuery, state: FSMContext):
    """Пагинация списка моделей"""
    _, provider, page_str = callback.data.split(":")
    page = int(page_str)
    
    try:
        await callback.message.edit_reply_markup(reply_markup=get_model_keyboard(provider, page=page))
    except Exception:
        # Если сообщение не изменилось, просто отвечаем на callback
        pass
    await callback.answer()


@router.callback_query(F.data.startswith("model:select:"))
async def process_model_selection(callback: CallbackQuery, state: FSMContext):
    """Выбор модели"""
    model = callback.data.split(":")[2]
    
    await state.update_data(model=model)
    
    await callback.message.edit_text(
        "🔑 <b>Введите API ключ</b>\n\n"
        "Введите ваш API ключ для выбранной модели:\n\n"
        "<i>Ваш ключ будет использоваться только для этой задачи</i>",
        reply_markup=get_cancel_keyboard()
    )
    await state.set_state(TaskState.waiting_for_api_key)


@router.message(TaskState.waiting_for_api_key)
async def process_api_key(message: Message, state: FSMContext):
    """Получение API ключа"""
    api_key = message.text.strip()
    
    if len(api_key) < 10:
        await message.answer("❌ API ключ слишком короткий. Проверьте его и попробуйте снова:")
        return
    
    await state.update_data(api_key=api_key)
    
    await message.answer(
        "📝 <b>Введите название задачи</b>\n\n"
        "Краткое название для вашей задачи:",
        reply_markup=get_cancel_keyboard()
    )
    await state.set_state(TaskState.waiting_for_title)


@router.message(TaskState.waiting_for_title)
async def process_title(message: Message, state: FSMContext):
    """Получение названия задачи"""
    title = message.text.strip()
    
    if len(title) < 3:
        await message.answer("❌ Название слишком короткое. Введите более подробное название:")
        return
    
    await state.update_data(title=title)
    
    await message.answer(
        "📋 <b>Опишите задачу</b>\n\n"
        "Подробно опишите, что нужно сделать:\n\n"
        "<i>Чем подробнее описание, тем лучше результат</i>",
        reply_markup=get_cancel_keyboard()
    )
    await state.set_state(TaskState.waiting_for_task)


@router.message(TaskState.waiting_for_task)
async def process_task(message: Message, state: FSMContext):
    """Получение описания задачи и создание задачи"""
    description = message.text.strip()
    
    if len(description) < 10:
        await message.answer("❌ Описание слишком короткое. Пожалуйста, опишите задачу подробнее:")
        return
    
    data = await state.get_data()
    token = data.get("token")
    title = data.get("title")
    provider = data.get("provider")
    model = data.get("model")
    api_key = data.get("api_key")
    workflow = data.get("workflow")
    username = data.get("username")  # Сохраняем имя пользователя
    
    # Создаём задачу через WebSocket
    loading_msg = await message.answer("⏳ <b>Подключаюсь к сервису...</b>")
    
    try:
        # Callback для обработки обновлений
        async def on_progress(update):
            """Обработка обновления прогресса"""
            msg_type = update.get("type", "")
            msg_text = update.get("message", "")
            progress = update.get("progress", 0)
            
            if msg_type == "connected":
                await message.answer(f"✅ Подключено! Task ID: {update.get('task_id', 'unknown')}")
            elif msg_type == "progress" or msg_type == "processing":
                # Отправляем обновления каждые 10% или важные сообщения
                if progress % 10 == 0 or "Boss" in msg_text or "Manager" in msg_text:
                    await message.answer(f"📊 Прогресс: {progress}%\n{msg_text}")
            elif msg_type == "success":
                zip_url = update.get("data", {}).get("zipUrl", "")
                files_count = update.get("data", {}).get("filesCount", 0)
                zip_size = update.get("data", {}).get("zipSize", 0)
                
                size_str = ""
                if zip_size:
                    size_kb = zip_size / 1024
                    size_str = f"\n📦 Размер: {size_kb:.1f} KB"
                
                await message.answer(
                    f"🎉 <b>Задача завершена!</b>\n\n"
                    f"✅ Файлов: {files_count}{size_str}\n\n"
                    f"📥 Скачать результат можно по ссылке:\n"
                    f"{zip_url if zip_url else 'Проверьте в веб-интерфейсе'}\n\n"
                    f"Хотите создать ещё одну задачу? Нажмите /start",
                )
            elif msg_type == "error":
                await message.answer(f"❌ <b>Ошибка выполнения</b>\n\n{msg_text}")
        
        # Запускаем создание задачи с WebSocket стримингом
        try:
            await loading_msg.delete()
        except Exception:
            pass  # Игнорируем ошибку удаления
        
        status_msg = await message.answer("🚀 <b>Задача запущена!</b>\n\nОжидайте обновлений...")
        
        async for update in api_service.create_task_and_stream(
            token=token,
            title=title,
            description=description,
            provider=provider,
            model=model,
            api_key=api_key,
            workflow=workflow,
            progress_callback=on_progress
        ):
            # updates уже обработаны через callback
            pass
        
        # Получаем сообщение о завершении
        await message.answer(
            "✅ <b>Готово!</b> Хотите создать ещё одну задачу? Нажмите /start"
        )
        
        # НЕ очищаем состояние полностью! 
        # Только удаляем данные задачи, сохраняя авторизацию
        await state.update_data(
            workflow_id=None,
            workflow=None,
            provider=None,
            model=None,
            api_key=None,
            title=None,
            description=None
        )
        await state.set_state(TaskState.selecting_workflow)
        
    except Exception as e:
        try:
            await loading_msg.delete()
        except Exception:
            pass
        await message.answer(f"❌ Ошибка создания задачи: {str(e)}")
        await state.clear()


@router.callback_query(F.data == "cancel")
async def cancel_task(callback: CallbackQuery, state: FSMContext):
    """Отмена создания задачи"""
    await state.clear()
    await callback.message.edit_text(
        "❌ Создание задачи отменено.\n\n"
        "Если хотите начать заново, нажмите /start"
    )

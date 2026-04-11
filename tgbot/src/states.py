"""
FSM состояния для бота
"""
from aiogram.fsm.state import State, StatesGroup


class AuthState(StatesGroup):
    """Состояния для процесса авторизации"""
    waiting_for_email = State()
    waiting_for_password = State()


class TaskState(StatesGroup):
    """Состояния для процесса создания задачи"""
    selecting_workflow = State()
    selecting_provider = State()
    selecting_model = State()
    waiting_for_api_key = State()
    waiting_for_task = State()
    waiting_for_title = State()

const WebSocket = require('ws');

const WS_URL = 'ws://localhost:3111/task/create';

const TASK_REQUEST = {
  username: "pavel",
  title: "Тест",
  description: "Напиши HTTP сервер на Go",
  tokens: {},
  meta: {
    provider: "cliproxy",
    model: "qwen-code"
  }
};

const RECONNECT_DELAY = 1000; // 1 секунда
const MAX_RECONNECT_DELAY = 10000; // 10 секунд максимум
let reconnectAttempts = 0;
let isTaskCompleted = false; // Флаг завершения задачи

function connect() {
  const delay = Math.min(RECONNECT_DELAY * Math.pow(2, reconnectAttempts), MAX_RECONNECT_DELAY);
  reconnectAttempts++;

  console.log(`🔌 Подключение к ${WS_URL}... (попытка ${reconnectAttempts})`);

  const ws = new WebSocket(WS_URL);

  ws.on('open', () => {
    console.log('✅ Соединение установлено');
    reconnectAttempts = 0; // Сбрасываем счётчик при успешном подключении

    console.log('📤 Отправляю запрос задачи...\n');
    ws.send(JSON.stringify(TASK_REQUEST));
  });

  ws.on('message', (data) => {
    let msg;
    try {
      msg = JSON.parse(data.toString());
    } catch {
      console.log('📨 Не-JSON сообщение:', data.toString());
      return;
    }

    const type = msg.type || 'unknown';
    const emoji = {
      connected: '🟢',
      processing: '🔄',
      progress: '🔄',
      error: '❌',
      done: '✅',
      success: '✅',
      status: 'ℹ️ ',
    }[type] || '📨';

    console.log(`${emoji} [${type}]`, JSON.stringify(msg, null, 2));
    console.log('---');

    if (type === 'error' || type === 'done' || type === 'success') {
      console.log('\n🏁 Завершение');
      isTaskCompleted = true;
      ws.close();
    }
  });

  ws.on('close', (code, reason) => {
    const reasonStr = reason ? `, reason: ${reason}` : '';
    console.log(`🔌 Соединение закрыто (code: ${code}${reasonStr})`);

    if (!isTaskCompleted) {
      console.log(`⏳ Переподключение через ${delay / 1000}с...`);
      setTimeout(connect, delay);
    } else {
      process.exit(0);
    }
  });

  ws.on('error', (err) => {
    console.error(`💥 Ошибка соединения: ${err.message}`);
    // WebSocket сам закроет соединение после ошибки, сработает on('close')
  });
}

// Запуск
connect();

// Глобальный таймаут 5 минут
setTimeout(() => {
  console.log('⏱️ Таймаут 5 минут, закрываю...');
  isTaskCompleted = true;
  process.exit(0);
}, 5 * 60 * 1000);

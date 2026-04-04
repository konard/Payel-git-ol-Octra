const WebSocket = require('ws');


const WS_URL = 'ws://localhost:3111/task/create'; 

const TASK_REQUEST = {
  username: "pavel",
  title: "Прокси сервер",
  description: "Напиши прокси сервис на go",
  tokens: {
    openrouter: "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"
  },
  meta: {
    provider: "openrouter",
    model: "qwen/qwen3.6-plus:free"
  }
};

console.log(`🔌 Подключение к ${WS_URL}...`);

const ws = new WebSocket(WS_URL);

ws.on('open', () => {
  console.log('✅ Соединение установлено');
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
    progress: '🔄',
    error: '❌',
    done: '✅',
    status: 'ℹ️ ',
  }[type] || '📨';

  console.log(`${emoji} [${type}]`, JSON.stringify(msg, null, 2));
  console.log('---');

  if (type === 'error' || type === 'done') {
    console.log('\n🏁 Завершение');
    ws.close();
  }
});

ws.on('close', (code) => {
  console.log(`🔌 Соединение закрыто (code: ${code})`);
  process.exit(0);
});

ws.on('error', (err) => {
  console.error('💥 Ошибка:', err.message);
  process.exit(1);
});

while (true) {
  setTimeout(() => {
    console.log('⏱️ Таймаут 5 минут, закрываю...');
    ws.close();
  }, 500 * 60 * 1000);
}

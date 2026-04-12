import { motion } from 'motion/react';

interface Setting {
  title: string;
  description: string;
  icon: string;
}

const settings: Setting[] = [
  {
    title: 'Параметры агентов',
    description: 'Настраивайте модель, температуру, API ключи для каждого агента',
    icon: '⚙',
  },
  {
    title: 'Тема оформления',
    description: 'Переключайтесь между тёмной и светлой темами',
    icon: '◐',
  },
  {
    title: 'Язык интерфейса',
    description: 'Поддержка 30+ языков с автоопределением',
    icon: 'Aa',
  },
  {
    title: 'Горячие клавиши',
    description: 'Настраиваемые shortcuts для быстрой работы',
    icon: '⌨',
  },
  {
    title: 'Профиль пользователя',
    description: 'Управление аккаунтом и подпиской',
    icon: '👤',
  },
  {
    title: 'Консоль отладки',
    description: 'Мониторинг выполнения задач',
    icon: '⊿',
  },
];

export function SettingsSection() {
  return (
    <section id="settings" className="py-20 px-4 sm:px-6 lg:px-8 bg-[var(--surface)]">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="text-center mb-16">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="inline-flex items-center gap-2 px-4 py-2 bg-[var(--accent)]/10 border border-[var(--accent)]/20 rounded-full text-sm font-medium text-[var(--accent)] mb-6"
          >
            Настройки
          </motion.div>

          <motion.h2
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="text-4xl sm:text-5xl font-bold text-[var(--text)] mb-6"
          >
            Полная
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-orange-500 to-orange-600"> кастомизация</span>
          </motion.h2>

          <motion.p
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="text-xl text-[var(--text-muted)] max-w-3xl mx-auto"
          >
            Настройте всё под себя — от внешнего вида до параметров каждого агента
          </motion.p>
        </div>

        {/* Settings grid */}
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {settings.map((setting, index) => (
            <motion.div
              key={setting.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: index * 0.05 }}
              className="p-6 bg-[var(--background)] border border-[var(--border)] rounded-xl hover:shadow-lg transition-all"
            >
              <div className="text-3xl mb-4">{setting.icon}</div>
              <h3 className="text-lg font-semibold text-[var(--text)] mb-2">
                {setting.title}
              </h3>
              <p className="text-sm text-[var(--text-muted)]">
                {setting.description}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

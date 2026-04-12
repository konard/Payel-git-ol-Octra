import { motion } from 'motion/react';

interface Feature {
  title: string;
  description: string;
  icon: string;
  color: string;
}

const features: Feature[] = [
  {
    title: 'Визуальный редактор',
    description: 'Создавайте и настраивайте workflow с помощью drag-and-drop интерфейса',
    icon: '⊞',
    color: '#ff6d5a',
  },
  {
    title: 'Иерархия агентов',
    description: 'Организуйте агентов в структуры Boss → Manager → Worker для эффективного управления',
    icon: '⊿',
    color: '#5a9bff',
  },
  {
    title: 'Real-time мониторинг',
    description: 'Наблюдайте за выполнением задач в реальном времени с анимацией потока данных',
    icon: '◉',
    color: '#50e3c2',
  },
  {
    title: 'Гибкая настройка',
    description: 'Настраивайте параметры каждого агента независимо от остальных',
    icon: '⚙',
    color: '#d97706',
  },
  {
    title: 'Библиотека шаблонов',
    description: 'Сохраняйте и переиспользуйте готовые workflow для разных задач',
    icon: '⊡',
    color: '#8b5cf6',
  },
  {
    title: 'Экспорт и импорт',
    description: 'Экспортируйте проекты в JSON и импортируйте их в любой момент',
    icon: '⇄',
    color: '#ec4899',
  },
  {
    title: 'Множество провайдеров',
    description: 'Поддерживаются OpenAI, Google, Anthropic, OpenRouter и многие другие',
    icon: '◎',
    color: '#06b6d4',
  },
  {
    title: 'Масштабируемость',
    description: 'От простых задач до сложных multi-agent систем без ограничений',
    icon: '⬡',
    color: '#f59e0b',
  },
];

export function FeaturesSection() {
  return (
    <section id="features" className="features-section">
      <div className="features-section__content">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="features-section__badge"
        >
          Возможности
        </motion.div>

        <motion.h2
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.1 }}
          className="features-section__title"
        >
          Всё что нужно для
          <span className="features-section__title-accent"> работы с AI</span>
        </motion.h2>

        <motion.p
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="features-section__description"
        >
          Полный набор инструментов для создания, управления и мониторинга multi-agent систем
        </motion.p>

        <div className="features-section__grid">
          {features.map((feature, index) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: index * 0.05 }}
              className="feature-card"
              whileHover={{ y: -5, transition: { duration: 0.2 } }}
            >
              <div
                className="feature-card__icon"
                style={{ color: feature.color }}
              >
                {feature.icon}
              </div>
              <h3 className="feature-card__title">{feature.title}</h3>
              <p className="feature-card__description">{feature.description}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

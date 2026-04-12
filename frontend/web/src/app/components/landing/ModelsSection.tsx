import { motion } from 'motion/react';
import agentsImage from '../../../images/main/dark/agents/agents.png';

interface ModelCategory {
  name: string;
  icon: string;
  models: string[];
  color: string;
}

const modelCategories: ModelCategory[] = [
  {
    name: 'Языковые модели',
    icon: 'Aa',
    models: ['GPT-4 Turbo', 'Claude 3 Opus', 'Gemini Ultra', 'Mistral Large', 'Llama 3 70B'],
    color: 'from-orange-500 to-orange-600',
  },
  {
    name: 'Модели для кода',
    icon: '</>',
    models: ['GPT-4 Code', 'CodeLlama', 'CodeGeeX', 'StarCoder', 'DeepSeek Coder'],
    color: 'from-blue-500 to-blue-600',
  },
  {
    name: 'Визуальные модели',
    icon: '🖼',
    models: ['DALL-E 3', 'Midjourney', 'Stable Diffusion', 'Flux', 'Ideogram'],
    color: 'from-green-500 to-green-600',
  },
  {
    name: 'Специализированные',
    icon: '⚙',
    models: ['Embeddings', 'Speech-to-Text', 'Text-to-Speech', 'Moderation', 'Translation'],
    color: 'from-purple-500 to-purple-600',
  },
];

export function ModelsSection() {
  return (
    <section id="models" className="py-20 px-4 sm:px-6 lg:px-8">
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
            Выбор моделей
          </motion.div>

          <motion.h2
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="text-4xl sm:text-5xl font-bold text-[var(--text)] mb-6"
          >
            Любые модели для
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-orange-500 to-orange-600"> любых задач</span>
          </motion.h2>

          <motion.p
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="text-xl text-[var(--text-muted)] max-w-3xl mx-auto"
          >
            Выбирайте из сотен доступных моделей для каждой задачи.
            Оптимизируйте стоимость и производительность.
          </motion.p>
        </div>

        {/* Agents image */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="mb-16 relative rounded-xl overflow-hidden shadow-2xl border border-[var(--border)] bg-[var(--surface)]"
        >
          <img
            src={agentsImage}
            alt="AI Models and Agents"
            className="w-full h-auto"
          />
        </motion.div>

        {/* Models grid */}
        <div className="grid md:grid-cols-2 gap-6">
          {modelCategories.map((category, index) => (
            <motion.div
              key={category.name}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: index * 0.1 }}
              className="p-6 bg-[var(--surface)] border border-[var(--border)] rounded-xl hover:shadow-lg transition-all"
            >
              <div className={`inline-flex items-center justify-center w-12 h-12 rounded-lg bg-gradient-to-br ${category.color} text-white font-bold text-lg mb-4`}>
                {category.icon}
              </div>
              <h3 className="text-xl font-semibold text-[var(--text)] mb-4">
                {category.name}
              </h3>
              <div className="space-y-2">
                {category.models.map((model) => (
                  <div
                    key={model}
                    className="px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-sm text-[var(--text-muted)]"
                  >
                    {model}
                  </div>
                ))}
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

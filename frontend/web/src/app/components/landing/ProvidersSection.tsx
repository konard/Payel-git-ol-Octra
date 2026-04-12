import { motion } from 'motion/react';
import { useThemeStore } from '../../../stores/themeStore';
import openaiIcon from '../../../images/icon.png';
import geminiIcon from '../../../images/gemini-color.png';
import claudeIcon from '../../../images/Claude_AI_symbol.svg';
import openrouterIcon from '../../../images/openrouter.svg';
import zaiIcon from '../../../images/zai.png';
import grokIcon from '../../../images/grok.png';
import qwenIcon from '../../../images/qwen-color.png';
import deepseekIcon from '../../../images/deepseek-color.png';
import providersDark from '../../../images/main/dark/providers/providers.png';
import providersLight from '../../../images/main/light/providers/providers.png';

interface Provider {
  name: string;
  icon: string;
}

const providers: Provider[] = [
  { name: 'OpenAI', icon: openaiIcon },
  { name: 'Google Gemini', icon: geminiIcon },
  { name: 'Anthropic Claude', icon: claudeIcon },
  { name: 'OpenRouter', icon: openrouterIcon },
  { name: 'Zhipu AI', icon: zaiIcon },
  { name: 'Grok', icon: grokIcon },
  { name: 'Qwen', icon: qwenIcon },
  { name: 'DeepSeek', icon: deepseekIcon },
];

export function ProvidersSection() {
  const { isDark } = useThemeStore();

  return (
    <section id="providers" className="py-20 px-4 sm:px-6 lg:px-8 bg-[var(--surface)]">
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
            Поддерживаемые провайдеры
          </motion.div>

          <motion.h2
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="text-4xl sm:text-5xl font-bold text-[var(--text)] mb-6"
          >
            Все популярные
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-orange-500 to-orange-600"> провайдеры</span>
          </motion.h2>

          <motion.p
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="text-xl text-[var(--text-muted)] max-w-3xl mx-auto"
          >
            Подключайте любые AI-модели от ведущих провайдеров.
            Используйте один API ключ для всех задач или комбинируйте несколько провайдеров.
          </motion.p>
        </div>

        {/* Providers grid */}
        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-16">
          {providers.map((provider, index) => (
            <motion.div
              key={provider.name}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: index * 0.05 }}
              className="p-6 bg-[var(--background)] border border-[var(--border)] rounded-xl hover:shadow-lg transition-all hover:-translate-y-1 flex flex-col items-center"
            >
              <div className="w-16 h-16 mb-4 rounded-lg flex items-center justify-center overflow-hidden">
                <img
                  src={provider.icon}
                  alt={provider.name}
                  className="w-full h-full object-contain"
                />
              </div>
              <h3 className="text-base font-semibold text-[var(--text)] text-center">
                {provider.name}
              </h3>
            </motion.div>
          ))}
        </div>

        {/* Providers interface image */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="relative rounded-xl overflow-hidden shadow-2xl border border-[var(--border)] bg-[var(--background)]"
        >
          <img
            src={isDark ? providersDark : providersLight}
            alt="Providers Interface"
            className="w-full h-auto"
          />
        </motion.div>
      </div>
    </section>
  );
}

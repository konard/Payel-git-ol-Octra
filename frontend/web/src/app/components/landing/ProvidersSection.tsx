import { motion } from 'motion/react';
import openaiIcon from '../../../images/icon.png';
import geminiIcon from '../../../images/gemini-color.png';
import claudeIcon from '../../../images/Claude_AI_symbol.svg';
import openrouterIcon from '../../../images/openrouter.svg';
import zaiIcon from '../../../images/zai.png';
import grokIcon from '../../../images/grok.png';
import qwenIcon from '../../../images/qwen-color.png';
import deepseekIcon from '../../../images/deepseek-color.png';

const providers = [
  { name: 'OpenAI', icon: openaiIcon },
  { name: 'Gemini', icon: geminiIcon },
  { name: 'Claude', icon: claudeIcon },
  { name: 'OpenRouter', icon: openrouterIcon },
  { name: 'Zhipu AI', icon: zaiIcon },
  { name: 'Grok', icon: grokIcon },
  { name: 'Qwen', icon: qwenIcon },
  { name: 'DeepSeek', icon: deepseekIcon },
];

export function ProvidersSection() {
  return (
    <section id="providers" className="bg-[#050505] px-4 py-20 text-white sm:px-6 lg:px-8">
      <div className="mx-auto max-w-7xl">
        <div className="grid gap-10 lg:grid-cols-[0.9fr_1.1fr] lg:items-center">
          <div>
            <p className="mb-4 text-sm font-semibold text-emerald-300">Model choice stays flexible</p>
            <h2 className="font-serif text-4xl font-semibold leading-tight tracking-tight sm:text-5xl">
              Use the provider that fits the task.
            </h2>
            <p className="mt-5 text-lg leading-8 text-white/66">
              A developer task can use a coding model, a research task can use a search-oriented setup,
              and document work can use the model your team already trusts.
            </p>
          </div>

          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            {providers.map((provider, index) => (
              <motion.div
                key={provider.name}
                initial={{ opacity: 0, y: 16 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.3, delay: index * 0.035 }}
                className="rounded-lg border border-white/10 bg-white/[0.035] p-4"
              >
                <img src={provider.icon} alt={provider.name} className="mb-4 h-10 w-10 object-contain" />
                <h3 className="text-sm font-semibold text-white">{provider.name}</h3>
              </motion.div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

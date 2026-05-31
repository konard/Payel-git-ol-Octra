import { Code2, FileText, Presentation, Search, Users } from 'lucide-react';
import { motion } from 'motion/react';

const audiences = [
  {
    title: 'Developers',
    description: 'Plan features, generate files, review implementation steps, and publish the resulting codebase.',
    icon: Code2,
  },
  {
    title: 'Researchers',
    description: 'Ask for topic briefs, compare sources, and turn search work into a concise written answer.',
    icon: Search,
  },
  {
    title: 'Document users',
    description: 'Create reports, summaries, specifications, and working notes without dealing with code views.',
    icon: FileText,
  },
  {
    title: 'Presentation teams',
    description: 'Draft slide structure and PPTX outputs for updates, classes, pitches, and internal reviews.',
    icon: Presentation,
  },
];

export function AgentsSection() {
  return (
    <section id="audiences" className="bg-[#0b0b0c] px-4 py-20 text-white sm:px-6 lg:px-8">
      <div className="mx-auto max-w-7xl">
        <div className="grid gap-10 lg:grid-cols-[0.8fr_1.2fr] lg:items-end">
          <div>
            <div className="mb-5 flex h-12 w-12 items-center justify-center rounded-lg border border-white/10 bg-white/7">
              <Users className="h-6 w-6 text-cyan-300" />
            </div>
            <h2 className="text-4xl font-semibold leading-tight sm:text-5xl">
              Regular users and technical teams work in the same system.
            </h2>
          </div>
          <p className="max-w-2xl text-lg leading-8 text-white/66 lg:justify-self-end">
            Octra is still useful for developers, but the landing page now reflects the broader product:
            regular users can ask for research, text documents, and presentations without translating
            their request into engineering language.
          </p>
        </div>

        <div className="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {audiences.map((audience, index) => {
            const Icon = audience.icon;

            return (
              <motion.article
                key={audience.title}
                initial={{ opacity: 0, y: 18 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.35, delay: index * 0.04 }}
                className="rounded-lg border border-white/10 bg-white/[0.035] p-5"
              >
                <Icon className="mb-8 h-6 w-6 text-orange-300" />
                <h3 className="text-lg font-semibold text-white">{audience.title}</h3>
                <p className="mt-3 text-sm leading-6 text-white/62">{audience.description}</p>
              </motion.article>
            );
          })}
        </div>
      </div>
    </section>
  );
}

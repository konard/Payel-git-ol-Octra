import { useEffect, useMemo, useRef, useState } from 'react';
import { CheckCircle2, Code2, FileText, Presentation, Search, type LucideIcon } from 'lucide-react';
import { motion } from 'motion/react';
import codeView from '../../../images/main/landing/code-view-dark.png';
import documentReader from '../../../images/main/landing/document-reader-dark.png';
import presentationDeck from '../../../images/main/landing/presentation-deck-dark.png';
import researchProgress from '../../../images/main/landing/research-progress-dark.png';

type ShowcaseItem = {
  id: string;
  eyebrow: string;
  title: string;
  description: string;
  image: string;
  icon: LucideIcon;
  points: string[];
};

const showcaseItems: ShowcaseItem[] = [
  {
    id: 'developers',
    eyebrow: 'For developers',
    title: 'Generate code with the same workspace your team reviews.',
    description:
      'Octra plans the work, assigns agents, streams files into the code viewer, and keeps the final project ready for a repository handoff.',
    image: codeView,
    icon: Code2,
    points: ['Code generation', 'Review-ready files', 'GitHub handoff'],
  },
  {
    id: 'research',
    eyebrow: 'For research',
    title: 'Turn open questions into sourced research summaries.',
    description:
      'Research workers search, compare, and consolidate findings so regular users can ask for a market scan, topic brief, or competitive note without building a workflow by hand.',
    image: researchProgress,
    icon: Search,
    points: ['Search progress', 'Fact gathering', 'Readable summaries'],
  },
  {
    id: 'documents',
    eyebrow: 'For text documents',
    title: 'Create readable reports, briefs, and working documents.',
    description:
      'Document tasks open in a focused reader instead of a code-first file tree, making the output easier to scan, page through, and share.',
    image: documentReader,
    icon: FileText,
    points: ['Markdown reports', 'Reader mode', 'Business-ready structure'],
  },
  {
    id: 'presentations',
    eyebrow: 'For presentations',
    title: 'Produce slide decks next to the written source.',
    description:
      'Octra can assemble presentation output for a pitch, class, or internal update while preserving the same agent trace used for code and research tasks.',
    image: presentationDeck,
    icon: Presentation,
    points: ['PPTX output', 'Slide outlines', 'Reusable source'],
  },
];

export function WorkflowDemo() {
  const [activeId, setActiveId] = useState(showcaseItems[0].id);
  const itemRefs = useRef<Record<string, HTMLDivElement | null>>({});

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((entry) => entry.isIntersecting)
          .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];

        const nextId = visible?.target.getAttribute('data-showcase-id');
        if (nextId) {
          setActiveId(nextId);
        }
      },
      {
        threshold: [0.35, 0.55, 0.75],
        rootMargin: '-18% 0px -34% 0px',
      },
    );

    showcaseItems.forEach((item) => {
      const node = itemRefs.current[item.id];
      if (node) observer.observe(node);
    });

    return () => observer.disconnect();
  }, []);

  const activeIndex = Math.max(0, showcaseItems.findIndex((item) => item.id === activeId));
  const activeItem = showcaseItems[activeIndex];

  const activeProgress = useMemo(
    () => ((activeIndex + 1) / showcaseItems.length) * 100,
    [activeIndex],
  );

  return (
    <section id="showcase" className="bg-[#050505] px-4 py-20 text-white sm:px-6 lg:px-8 lg:py-28">
      <div className="mx-auto max-w-7xl">
        <div className="mb-14 max-w-3xl">
          <p className="mb-4 text-sm font-semibold text-orange-300">One workspace, multiple outcomes</p>
          <h2 className="font-serif text-4xl font-semibold leading-tight tracking-tight sm:text-5xl">
            Built for code, research, text documents, and presentations.
          </h2>
          <p className="mt-5 text-lg leading-8 text-white/66">
            The page stays calm while each example moves through the same Octra flow:
            describe the task, let agents work, then inspect the finished result.
          </p>
        </div>

        <div className="grid gap-12 lg:grid-cols-[minmax(0,0.88fr)_minmax(0,1.12fr)] lg:gap-16">
          <div className="space-y-8 lg:space-y-[18vh] lg:pb-[14vh]">
            {showcaseItems.map((item, index) => {
              const Icon = item.icon;
              const isActive = item.id === activeId;

              return (
                <motion.article
                  key={item.id}
                  ref={(node) => {
                    itemRefs.current[item.id] = node;
                  }}
                  data-showcase-id={item.id}
                  initial={{ opacity: 0, y: 24 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true, margin: '-15% 0px' }}
                  transition={{ duration: 0.45 }}
                  className={`rounded-lg border p-6 transition-colors lg:min-h-[42vh] lg:p-8 ${
                    isActive
                      ? 'border-orange-400/70 bg-white/[0.055]'
                      : 'border-white/10 bg-white/[0.028]'
                  }`}
                >
                  <div className="mb-6 flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg border border-white/12 bg-white/8">
                      <Icon className="h-5 w-5 text-orange-300" />
                    </div>
                    <div>
                      <p className="text-sm font-semibold text-orange-300">{item.eyebrow}</p>
                      <p className="text-sm text-white/44">{String(index + 1).padStart(2, '0')} / 04</p>
                    </div>
                  </div>

                  <h3 className="text-2xl font-semibold leading-tight text-white sm:text-3xl">
                    {item.title}
                  </h3>
                  <p className="mt-5 text-base leading-7 text-white/66">{item.description}</p>

                  <ul className="mt-7 grid gap-3">
                    {item.points.map((point) => (
                      <li key={point} className="flex items-center gap-3 text-sm font-medium text-white/78">
                        <CheckCircle2 className="h-4 w-4 text-emerald-300" />
                        {point}
                      </li>
                    ))}
                  </ul>

                  <img
                    src={item.image}
                    alt={`${item.eyebrow} screenshot`}
                    className="mt-7 block w-full rounded-lg border border-white/10 bg-black lg:hidden"
                    loading="lazy"
                  />
                </motion.article>
              );
            })}
          </div>

          <div className="hidden lg:block">
            <div className="sticky top-24">
              <div className="overflow-hidden rounded-lg border border-white/10 bg-[#111113] shadow-[0_32px_90px_rgba(0,0,0,0.48)]">
                <div className="flex h-12 items-center justify-between border-b border-white/10 px-4">
                  <div className="flex items-center gap-2">
                    <span className="h-3 w-3 rounded-full bg-[#ff5f57]" />
                    <span className="h-3 w-3 rounded-full bg-[#ffbd2e]" />
                    <span className="h-3 w-3 rounded-full bg-[#28c840]" />
                  </div>
                  <span className="text-sm font-medium text-white/58">{activeItem.eyebrow}</span>
                </div>

                <div className="relative aspect-[16/10] bg-black">
                  {showcaseItems.map((item) => (
                    <img
                      key={item.id}
                      src={item.image}
                      alt={`${item.eyebrow} screenshot`}
                      className={`absolute inset-0 h-full w-full object-cover transition-opacity duration-500 ${
                        item.id === activeId ? 'opacity-100' : 'opacity-0'
                      }`}
                    />
                  ))}
                </div>
              </div>

              <div className="mt-5 h-1 overflow-hidden rounded-full bg-white/10">
                <div
                  className="h-full rounded-full bg-orange-400 transition-[width] duration-500"
                  style={{ width: `${activeProgress}%` }}
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

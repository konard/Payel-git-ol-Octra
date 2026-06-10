import octraVideo from '../../../images/main/video/octra-animation-landing.mp4';
import { ArrowRight, FileText, Presentation, Search } from 'lucide-react';

export function HeroSection() {
  return (
    <section className="relative flex min-h-[86dvh] items-center overflow-hidden bg-black px-4 pb-14 pt-24 sm:px-6 lg:px-8">
      <div className="absolute inset-0 z-0">
        <video
          autoPlay
          loop
          muted
          playsInline
          className="absolute inset-0 w-full h-full object-cover"
        >
          <source src={octraVideo} type="video/mp4" />
        </video>
        <div className="absolute inset-0 bg-black/68" />
        <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(0,0,0,0.42)_0%,rgba(0,0,0,0.18)_42%,#050505_100%)]" />
      </div>

      <div className="relative z-10 mx-auto max-w-5xl pt-8 text-center">
        <div className="mb-6 flex flex-wrap items-center justify-center gap-3 text-sm font-medium text-white/70">
          <span className="inline-flex items-center gap-2">
            <Search className="h-4 w-4 text-cyan-300" />
            Research
          </span>
          <span className="h-1 w-1 rounded-full bg-white/34" />
          <span className="inline-flex items-center gap-2">
            <FileText className="h-4 w-4 text-emerald-300" />
            Text documents
          </span>
          <span className="h-1 w-1 rounded-full bg-white/34" />
          <span className="inline-flex items-center gap-2">
            <Presentation className="h-4 w-4 text-violet-300" />
            Presentations
          </span>
        </div>

        <h1 className="mb-6 font-serif text-5xl font-semibold leading-none tracking-tight text-white sm:text-6xl lg:text-7xl">
          Octra
        </h1>

        <p className="mx-auto mb-10 max-w-3xl text-xl leading-8 text-white/76">
          AI teams for developers, regular users, research, text documents, and slide decks.
          Describe the result you need and follow the work from first plan to finished output.
        </p>

        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <a
            href="/app"
            className="inline-flex items-center justify-center gap-2 rounded-md bg-white px-8 py-3 text-base font-semibold text-black transition-colors hover:bg-white/88"
          >
            Start a task <ArrowRight className="h-5 w-5" />
          </a>
          <a
            href="#showcase"
            className="inline-flex items-center justify-center rounded-md border border-white/20 bg-white/8 px-8 py-3 text-base font-semibold text-white transition-colors hover:bg-white/13"
          >
            View workflows
          </a>
        </div>
      </div>
    </section>
  );
}

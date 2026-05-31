import { useEffect } from 'react';
import { ArrowRight } from 'lucide-react';
import { HeroSection } from './landing/HeroSection';
import { WorkflowDemo } from './landing/WorkflowDemo';
import { ProvidersSection } from './landing/ProvidersSection';
import { AgentsSection } from './landing/AgentsSection';
import { FooterSection } from './landing/FooterSection';
import octraMascot from '../../images/octra-mascot.png';

export default function LandingPage() {
  useEffect(() => {
    window.scrollTo(0, 0);
    document.documentElement.classList.add('dark');
  }, []);

  const navItems = [
    { href: '#showcase', label: 'Showcase' },
    { href: '#audiences', label: 'Audiences' },
    { href: '#providers', label: 'Models' },
  ];

  return (
    <div className="min-h-screen bg-[#050505] text-white">
      <nav className="fixed top-0 left-0 right-0 z-50 border-b border-white/10 bg-black/78 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <img
                src={octraMascot}
                alt="Octra Mascot"
                className="w-9 h-9 rounded-md object-contain"
              />
              <span className="text-lg font-semibold">Octra</span>
            </div>

            <div className="hidden md:flex items-center gap-7">
              {navItems.map((item) => (
                <a
                  key={item.href}
                  href={item.href}
                  className="text-sm font-medium text-white/62 transition-colors hover:text-white"
                >
                  {item.label}
                </a>
              ))}
            </div>

            <div className="flex items-center gap-3">
              <a
                href="/app"
                className="hidden sm:inline-flex px-4 py-2 text-sm font-medium text-white/78 transition-colors hover:text-white"
              >
                Sign in
              </a>
              <a
                href="/app"
                className="inline-flex items-center gap-2 rounded-md bg-white px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-white/88"
              >
                Start <ArrowRight className="h-4 w-4" />
              </a>
            </div>
          </div>
        </div>
      </nav>

      <HeroSection />
      <WorkflowDemo />
      <AgentsSection />
      <ProvidersSection />
      <FooterSection />
    </div>
  );
}

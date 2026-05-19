import octraVideo from '../../../images/main/video/octra-animation-landing.mp4';
import { useI18n } from '../../../hooks/useI18n';

export function HeroSection() {
  const { t } = useI18n();
  return (
    <section className="relative min-h-[100dvh] flex items-center pt-20 pb-16 px-4 sm:px-6 lg:px-8 overflow-hidden bg-[#0a0a0f]">
      {/* Video Background - Raycast style */}
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
        {/* Dark overlay for text readability */}
        <div className="absolute inset-0 bg-black/60" />
        {/* Subtle gradient for depth */}
        <div className="absolute inset-0 bg-gradient-to-b from-black/40 via-black/30 to-black/70" />
      </div>

      <div className="max-w-5xl mx-auto relative z-10 text-center pt-16">
        {/* Badge */}
        <div className="inline-flex items-center gap-2 px-4 py-1.5 bg-white/10 backdrop-blur-md border border-white/20 rounded-full text-sm font-medium text-white/90 mb-8">
          <span className="w-1.5 h-1.5 bg-orange-400 rounded-full animate-pulse" />
          {t('landing.hero.badge')}
        </div>

        {/* Headline */}
        <h1 
          className="text-6xl sm:text-7xl lg:text-8xl font-semibold tracking-tighter text-white mb-6 leading-none"
          dangerouslySetInnerHTML={{ __html: t('landing.hero.title') }} 
        />

        <p className="text-xl text-white/70 max-w-xl mx-auto mb-10">
          {t('landing.hero.subtitle')}
        </p>

        {/* CTAs */}
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <a
            href="/app"
            className="inline-flex items-center justify-center gap-2 px-9 py-3.5 text-base font-semibold bg-white text-black rounded-xl hover:bg-white/90 active:scale-[0.985] transition-all"
          >
            {t('landing.hero.ctaStart')}
          </a>
          <a
            href="#workflow"
            className="inline-flex items-center justify-center gap-2 px-9 py-3.5 text-base font-semibold bg-white/10 text-white border border-white/20 hover:bg-white/15 rounded-xl transition-all"
          >
            {t('landing.hero.ctaDemo')}
          </a>
        </div>
      </div>
    </section>
  );
}


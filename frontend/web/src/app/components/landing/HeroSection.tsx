import { useThemeStore } from '../../../stores/themeStore';
import workflowDark from '../../../images/main/dark/workfloy/workflow_dark_theme_no_right_panel.png';
import workflowLight from '../../../images/main/light/main_screen.png';

export function HeroSection() {
  const { isDark } = useThemeStore();

  return (
    <section className="relative pt-32 pb-20 px-4 sm:px-6 lg:px-8 overflow-hidden">
      {/* Background gradient */}
      <div className="absolute inset-0 bg-gradient-to-b from-[var(--accent)]/5 via-transparent to-transparent pointer-events-none" />

      <div className="max-w-7xl mx-auto relative">
        {/* Text content */}
        <div className="text-center max-w-4xl mx-auto mb-12">
          <div className="inline-flex items-center gap-2 px-4 py-2 bg-[var(--accent)]/10 border border-[var(--accent)]/20 rounded-full text-sm font-medium text-[var(--accent)] mb-6">
            <span className="w-2 h-2 bg-[var(--accent)] rounded-full animate-pulse" />
            Визуальный редактор AI-агентов
          </div>

          <h1 className="text-5xl sm:text-6xl lg:text-7xl font-bold text-[var(--text)] mb-6 leading-tight">
            Создавайте мощные
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-orange-500 to-orange-600"> AI-команды</span>
            <br />визуально
          </h1>

          <p className="text-xl text-[var(--text-muted)] mb-8 max-w-3xl mx-auto">
            Управляйте множеством AI-агентов через интуитивный визуальный интерфейс.
            Создавайте workflow, подключайте любые модели и автоматизируйте задачи.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <a
              href="/app"
              className="inline-flex items-center justify-center gap-2 px-8 py-3 text-base font-semibold bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white rounded-lg transition-all shadow-lg hover:shadow-xl hover:-translate-y-0.5"
            >
              Начать бесплатно
              <span className="text-lg">→</span>
            </a>
            <a
              href="#workflow"
              className="inline-flex items-center justify-center gap-2 px-8 py-3 text-base font-semibold text-[var(--text)] bg-[var(--surface)] hover:bg-[var(--background)] border border-[var(--border)] rounded-lg transition-all"
            >
              Посмотреть демо
            </a>
          </div>
        </div>

        {/* Workflow screenshot preview */}
        <div className="relative max-w-5xl mx-auto">
          <div className="relative rounded-xl overflow-hidden shadow-2xl border border-[var(--border)] bg-[var(--surface)]">
            <img
              src={isDark ? workflowDark : workflowLight}
              alt="CrewAI Workflow Interface"
              className="w-full h-auto"
            />
            {/* Gradient overlay at bottom */}
            <div className="absolute inset-0 bg-gradient-to-t from-[var(--background)] via-transparent to-transparent opacity-20" />
          </div>
          
          {/* Decorative elements */}
          <div className="absolute -top-4 -left-4 w-24 h-24 bg-[var(--accent)]/10 rounded-full blur-2xl" />
          <div className="absolute -bottom-4 -right-4 w-32 h-32 bg-[var(--accent)]/10 rounded-full blur-2xl" />
        </div>
      </div>
    </section>
  );
}

import crewaiMascot from '../../../images/crewai-mascot.png';

export function FooterSection() {
  return (
    <footer className="border-t border-[var(--border)] bg-[var(--background)]">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid md:grid-cols-4 gap-8 mb-8">
          {/* Brand */}
          <div className="md:col-span-1">
            <div className="flex items-center gap-3 mb-4">
              <img
                src={crewaiMascot}
                alt="CrewAI Mascot"
                className="w-10 h-10 rounded-lg object-contain"
              />
              <span className="text-lg font-semibold text-[var(--text)]">CrewAI</span>
            </div>
            <p className="text-sm text-[var(--text-muted)]">
              Визуальный редактор для управления AI-агентами
            </p>
          </div>

          {/* Links */}
          <div>
            <h4 className="text-sm font-semibold text-[var(--text)] mb-4">Продукт</h4>
            <ul className="space-y-2">
              <li><a href="#workflow" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">Workflow</a></li>
              <li><a href="#agents" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">Агенты</a></li>
              <li><a href="#providers" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">Провайдеры</a></li>
              <li><a href="#models" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">Модели</a></li>
            </ul>
          </div>

          <div>
            <h4 className="text-sm font-semibold text-[var(--text)] mb-4">Ресурсы</h4>
            <ul className="space-y-2">
              <li><a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">Документация</a></li>
              <li><a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">Руководства</a></li>
              <li><a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">API Reference</a></li>
              <li><a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">Примеры</a></li>
            </ul>
          </div>

          <div>
            <h4 className="text-sm font-semibold text-[var(--text)] mb-4">Компания</h4>
            <ul className="space-y-2">
              <li><a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">О нас</a></li>
              <li><a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">Блог</a></li>
              <li><a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">Карьера</a></li>
              <li><a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">Контакты</a></li>
            </ul>
          </div>
        </div>

        {/* Bottom */}
        <div className="pt-8 border-t border-[var(--border)] flex flex-col md:flex-row justify-between items-center gap-4">
          <p className="text-sm text-[var(--text-muted)]">
            © {new Date().getFullYear()} CrewAI. Все права защищены.
          </p>
          <div className="flex gap-6">
            <a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">
              Политика конфиденциальности
            </a>
            <a href="#" className="text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">
              Условия использования
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}

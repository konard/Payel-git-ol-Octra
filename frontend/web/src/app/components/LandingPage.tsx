import { useState, useEffect } from 'react';
import { useThemeStore } from '../../stores/themeStore';
import { useAuthStore } from '../../stores/authStore';
import { HeroSection } from './landing/HeroSection';
import { WorkflowDemo } from './landing/WorkflowDemo';
import { ProvidersSection } from './landing/ProvidersSection';
import { AgentsSection } from './landing/AgentsSection';
import { ModelsSection } from './landing/ModelsSection';
import { SettingsSection } from './landing/SettingsSection';
import { FooterSection } from './landing/FooterSection';

export default function LandingPage() {
  const { isDark } = useThemeStore();
  const { isAuthenticated } = useAuthStore();

  useEffect(() => {
    // Scroll to top on mount
    window.scrollTo(0, 0);
  }, []);

  return (
    <div className="min-h-screen bg-[var(--background)] text-[var(--text)]">
      {/* Navigation */}
      <nav className="fixed top-0 left-0 right-0 z-50 bg-[var(--background)]/80 backdrop-blur-md border-b border-[var(--border)]">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            {/* Logo */}
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 bg-gradient-to-br from-orange-500 to-orange-600 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold text-sm">C</span>
              </div>
              <span className="text-lg font-semibold text-[var(--text)]">CrewAI</span>
            </div>

            {/* Links */}
            <div className="hidden md:flex items-center gap-8">
              <a href="#workflow" className="text-sm font-medium text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">
                Workflow
              </a>
              <a href="#agents" className="text-sm font-medium text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">
                Агенты
              </a>
              <a href="#providers" className="text-sm font-medium text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">
                Провайдеры
              </a>
              <a href="#models" className="text-sm font-medium text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">
                Модели
              </a>
              <a href="#settings" className="text-sm font-medium text-[var(--text-muted)] hover:text-[var(--text)] transition-colors">
                Настройки
              </a>
            </div>

            {/* Actions */}
            <div className="flex items-center gap-3">
              <a
                href="/app"
                className="px-4 py-2 text-sm font-medium text-[var(--text)] hover:bg-[var(--surface)] rounded-md transition-colors"
              >
                Войти
              </a>
              <a
                href="/app"
                className="px-4 py-2 text-sm font-medium bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white rounded-md transition-all shadow-md hover:shadow-lg"
              >
                Начать
              </a>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <HeroSection />

      {/* Workflow Demo Section */}
      <WorkflowDemo />

      {/* Agents Section */}
      <AgentsSection />

      {/* Providers Section */}
      <ProvidersSection />

      {/* Models Section */}
      <ModelsSection />

      {/* Settings Section */}
      <SettingsSection />

      {/* Footer */}
      <FooterSection />
    </div>
  );
}

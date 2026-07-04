'use client';

import type { ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { Bell, Plus, Trophy } from 'lucide-react';
import { useState } from 'react';
import { UserBalance } from '../components/UserBalance';
import { dashboardSections, dashboardTabs } from './sections';
import { ROUTES } from '../config/routes';
import { ASSETS } from '../config/images';
import { CreateEnvironmentModal } from '../components/CreateEnvironmentModal';
import { createDashboardEnvironment } from '../server/environments';

type DashboardShellProps = {
  activeSection: string;
  children: ReactNode;
  hideSidebarItems?: string[];
  showNotifications?: boolean;
  hideNewFlow?: boolean;
  hideTabs?: boolean;
};

export function DashboardShell({ activeSection, children, hideSidebarItems, showNotifications = true, hideNewFlow = false, hideTabs = false }: DashboardShellProps) {
  const router = useRouter();
  const [showCreate, setShowCreate] = useState(false);
  const [createError, setCreateError] = useState('');

  async function handleCreate(name: string, visibility: 'private' | 'public') {
    setCreateError('');
    const res = await createDashboardEnvironment(name, visibility);
    if (!res.ok) {
      const text = await res.text();
      setCreateError(text || 'Failed to create environment');
      return;
    }
    const env = await res.json();
    document.cookie = `octra_selected_env=${env.id}; path=/; max-age=31536000; SameSite=Lax`;
    setShowCreate(false);
    router.push(ROUTES.DASHBOARD_ENVIRONMENTS);
  }

  return (
    <main className="dashboard-page">
      <aside className="app-sidebar" aria-label="Octra sections">
        <a className="square-brand" href={ROUTES.WORKSPACE} aria-label="Octra home">
          <img src={ASSETS.LOGO} alt="" />
        </a>
        <nav>
          {dashboardSections.map((item) => {
            const Icon = item.icon;
            const hidden = hideSidebarItems ?? ['models', 'files', 'security', 'settings'];
            if (hidden.includes(item.slug)) return null;
            return (
              <a
                className={item.slug === activeSection ? 'side-icon active' : 'side-icon'}
                href={item.href}
                key={item.slug}
                aria-label={item.label}
              >
                <Icon size={18} />
              </a>
            );
          })}
        </nav>
      </aside>

      <section className="dashboard-scene" aria-labelledby="dashboard-title">
        <header className="dashboard-topbar">
          <div className="crumbs">
            <span id="dashboard-title">Pipeline Command</span>
          </div>
          <div className="topbar-actions">
            <UserBalance />
            <a className="icon-button dark-icon" href={ROUTES.PROFILE_LEADERBOARD} aria-label="User leaderboard" title="User leaderboard">
              <Trophy size={18} />
            </a>
            {showNotifications && (
              <a className="icon-button dark-icon" href={ROUTES.DASHBOARD_SECURITY} aria-label="Notifications">
                <Bell size={18} />
              </a>
            )}
            {!hideNewFlow && (
              <button className="small-command accent-command" onClick={() => { setCreateError(''); setShowCreate(true); }}>
                <Plus size={15} />
                New
              </button>
            )}
          </div>
        </header>

        {!hideTabs && (
          <div className="dashboard-tabs" aria-label="Dashboard views">
            {dashboardTabs.map((tab) => (
              <a className={tab.slug === activeSection ? 'active' : ''} href={tab.href} key={tab.slug}>
                {tab.label}
            </a>
          ))}
        </div>
        )}

        {children}
      </section>

      {showCreate && (
        <CreateEnvironmentModal
          onClose={() => setShowCreate(false)}
          onCreate={handleCreate}
          error={createError}
        />
      )}
    </main>
  );
}

import type { ReactNode } from 'react';
import { Bell, PanelLeft, Plus, Search } from 'lucide-react';
import { UserBalance } from '../components/UserBalance';
import { dashboardSections, dashboardTabs } from './sections';

type DashboardShellProps = {
  activeSection: string;
  children: ReactNode;
};

export function DashboardShell({ activeSection, children }: DashboardShellProps) {
  return (
    <main className="dashboard-page">
      <aside className="app-sidebar" aria-label="Octra sections">
        <a className="square-brand" href="/app" aria-label="Octra home">
          <img src="/assets/octra-node-logo.svg" alt="" />
        </a>
        <nav>
          {dashboardSections.map((item) => {
            const Icon = item.icon;
            // Models, Files, Security buttons are commented out by user request. To restore, remove the line below
            if (['models', 'files', 'security'].includes(item.slug)) return null;
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
            <a className="icon-button dark-icon" href="/app" aria-label="Open app workspace">
              <PanelLeft size={18} />
            </a>
            <a href="/">Octra</a>
            <span id="dashboard-title">Pipeline Command</span>
          </div>
          <label className="search-field dashboard-search">
            <Search size={15} />
            <span className="sr-only">Search dashboard</span>
            <input placeholder="Search..." />
          </label>
          <div className="topbar-actions">
            <UserBalance />
            <a className="icon-button dark-icon" href="/dashboard/security" aria-label="Notifications">
              <Bell size={18} />
            </a>
            <a className="small-command accent-command" href="/dashboard/flows">
              <Plus size={15} />
              New flow
            </a>
          </div>
        </header>

        <div className="dashboard-tabs" aria-label="Dashboard views">
          {dashboardTabs.map((tab) => (
            <a className={tab.slug === activeSection ? 'active' : ''} href={tab.href} key={tab.slug}>
              {tab.label}
            </a>
          ))}
        </div>

        {children}
      </section>
    </main>
  );
}

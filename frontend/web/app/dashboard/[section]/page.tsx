import { notFound } from 'next/navigation';
import { EmptyDataPanel } from '../../components/EmptyDataPanel';
import { EnvironmentPanel } from '../../components/EnvironmentPanel';
import { DashboardShell } from '../DashboardShell';
import { findDashboardSection, routeSections } from '../sections';
import { ROUTES } from '../../config/routes';

export const dynamicParams = false;

export function generateStaticParams() {
  return routeSections
    .filter((section) => section.slug !== 'overview' && section.slug !== 'settings' && section.slug !== 'environments')
    .map((section) => ({ section: section.slug }));
}

export default async function DashboardSectionPage({ params }: { params: Promise<{ section: string }> }) {
  const { section: slug } = await params;
  const section = findDashboardSection(slug);
  if (!section || section.slug === 'overview' || section.slug === 'settings') {
    notFound();
  }

  const Icon = section.icon;
  const isEnvironmentSection = section.slug === 'environments' || section.slug === 'flows';

  return (
    <DashboardShell activeSection={section.slug} hideTabs={isEnvironmentSection} showNotifications={!isEnvironmentSection}>
      <section className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label={section.title}>
          {isEnvironmentSection ? (
            <div className="dashboard-environment-panel">
              <div className="panel-heading">
                <span>All environments</span>
                <a className="ghost-command" href={ROUTES.WORKSPACE}>
                  Open workspace
                </a>
              </div>
              <EnvironmentPanel mode="manage" />
            </div>
          ) : (
            <EmptyDataPanel
              icon={Icon}
              title={section.title}
              detail={section.detail}
              actionHref={ROUTES.WORKSPACE}
              actionLabel="Open workspace"
            />
          )}
        </article>
      </section>
    </DashboardShell>
  );
}

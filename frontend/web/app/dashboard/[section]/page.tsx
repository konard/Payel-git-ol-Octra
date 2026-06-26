import { notFound } from 'next/navigation';
import { EmptyDataPanel } from '../../components/EmptyDataPanel';
import { DashboardShell } from '../DashboardShell';
import { findDashboardSection, routeSections } from '../sections';

export const dynamicParams = false;

export function generateStaticParams() {
  return routeSections
    .filter((section) => section.slug !== 'overview' && section.slug !== 'settings')
    .map((section) => ({ section: section.slug }));
}

export default async function DashboardSectionPage({ params }: { params: Promise<{ section: string }> }) {
  const { section: slug } = await params;
  const section = findDashboardSection(slug);
  if (!section || section.slug === 'overview' || section.slug === 'settings') {
    notFound();
  }

  const Icon = section.icon;

  return (
    <DashboardShell activeSection={section.slug}>
      <section className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label={section.title}>
          <EmptyDataPanel
            icon={Icon}
            title={section.title}
            detail={section.detail}
            actionHref="/app"
            actionLabel="Open workspace"
          />
        </article>
      </section>
    </DashboardShell>
  );
}

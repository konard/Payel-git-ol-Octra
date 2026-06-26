import {
  BarChart3,
  Bot,
  FileText,
  Gauge,
  GitBranch,
  LockKeyhole,
  Rocket,
  Settings,
  ShieldCheck,
  Workflow,
  type LucideIcon,
} from 'lucide-react';

export type DashboardSection = {
  slug: string;
  label: string;
  title: string;
  detail: string;
  href: string;
  icon: LucideIcon;
};

export const dashboardSections: DashboardSection[] = [
  {
    slug: 'overview',
    label: 'Overview',
    title: 'Overview',
    detail: 'Live workspace activity will appear here when environments start reporting telemetry.',
    href: '/dashboard',
    icon: Gauge,
  },
  {
    slug: 'flows',
    label: 'Flows',
    title: 'Flows',
    detail: 'Saved environment flows will appear here after they are created.',
    href: '/dashboard/flows',
    icon: Workflow,
  },
  {
    slug: 'models',
    label: 'Models',
    title: 'Models',
    detail: 'Configured model routes will appear here after provider settings are connected.',
    href: '/dashboard/models',
    icon: Bot,
  },
  {
    slug: 'files',
    label: 'Files',
    title: 'Files',
    detail: 'Generated files and workspace artifacts will appear here after a run produces output.',
    href: '/dashboard/files',
    icon: FileText,
  },
  {
    slug: 'security',
    label: 'Security',
    title: 'Security',
    detail: 'Access policies and audit events will appear here after security data is available.',
    href: '/dashboard/security',
    icon: ShieldCheck,
  },
  {
    slug: 'settings',
    label: 'Settings',
    title: 'Settings',
    detail: 'Account, billing, and workspace settings will be managed from this screen.',
    href: '/settings',
    icon: Settings,
  },
];

export const dashboardTabs: DashboardSection[] = [
  dashboardSections[0],
  {
    slug: 'metrics',
    label: 'Metrics',
    title: 'Metrics',
    detail: 'Runtime counters will appear here after telemetry is connected.',
    href: '/dashboard/metrics',
    icon: BarChart3,
  },
  {
    slug: 'evaluations',
    label: 'Evaluations',
    title: 'Evaluations',
    detail: 'Evaluation results will appear here after benchmark runs complete.',
    href: '/dashboard/evaluations',
    icon: GitBranch,
  },
  {
    slug: 'deployments',
    label: 'Deployments',
    title: 'Deployments',
    detail: 'Deployment status will appear here after environments publish releases.',
    href: '/dashboard/deployments',
    icon: Rocket,
  },
];

export const accountSection: DashboardSection = {
  slug: 'account',
  label: 'Account',
  title: 'Account',
  detail: 'Account identity and connected providers will appear here after sign-in.',
  href: '/login',
  icon: LockKeyhole,
};

export const routeSections = [...dashboardSections, ...dashboardTabs.filter((tab) => tab.slug !== 'overview')];

export function findDashboardSection(slug: string): DashboardSection | undefined {
  return routeSections.find((section) => section.slug === slug);
}

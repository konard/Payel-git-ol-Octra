import {
  BarChart3,
  Bot,
  Cable,
  FileText,
  Gauge,
  GitBranch,
  LockKeyhole,
  Play,
  Rocket,
  Settings,
  ShieldCheck,
  UserCheck,
  Workflow,
  Zap,
  type LucideIcon,
} from 'lucide-react';
import { ROUTES } from '../config/routes';


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
    href: ROUTES.DASHBOARD,
    icon: Gauge,
  },
  {
    slug: 'environments',
    label: 'Environments',
    title: 'Environments',
    detail: 'Saved environments will appear here after they are created.',
    href: ROUTES.DASHBOARD_ENVIRONMENTS,
    icon: Workflow,
  },
  {
    slug: 'mcp',
    label: 'MCP',
    title: 'MCP Servers',
    detail: 'Manage your MCP servers and browse available tools, resources, and prompts.',
    href: ROUTES.DASHBOARD_MCP,
    icon: Cable,
  },
  {
    slug: 'runs',
    label: 'Runs',
    title: 'Workflow Runs',
    detail: 'View and manage workflow runs across your environments.',
    href: ROUTES.DASHBOARD_RUNS,
    icon: Play,
  },
  {
    slug: 'hitl',
    label: 'HITL',
    title: 'Human-in-the-Loop',
    detail: 'Review and respond to workflow runs waiting for human input.',
    href: ROUTES.DASHBOARD_HITL,
    icon: UserCheck,
  },
  {
    slug: 'triggers',
    label: 'Triggers',
    title: 'Triggers & Webhooks',
    detail: 'HTTP-trigger endpoints for automating workflows, agents, and skills.',
    href: ROUTES.DASHBOARD_TRIGGERS,
    icon: Zap,
  },
  {
    slug: 'flows',
    label: 'Flows',
    title: 'Flows',
    detail: 'Workflow canvas runs will appear here after nodes are connected.',
    href: ROUTES.DASHBOARD_FLOWS,
    icon: Workflow,
  },
  {
    slug: 'models',
    label: 'Models',
    title: 'Models',
    detail: 'Configured model routes will appear here after provider settings are connected.',
    href: ROUTES.DASHBOARD_MODELS,
    icon: Bot,
  },
  {
    slug: 'files',
    label: 'Files',
    title: 'Files',
    detail: 'Generated files and workspace artifacts will appear here after a run produces output.',
    href: ROUTES.DASHBOARD_FILES,
    icon: FileText,
  },
  {
    slug: 'security',
    label: 'Security',
    title: 'Security',
    detail: 'Access policies and audit events will appear here after security data is available.',
    href: ROUTES.DASHBOARD_SECURITY,
    icon: ShieldCheck,
  },
  {
    slug: 'settings',
    label: 'Settings',
    title: 'Settings',
    detail: 'Account, billing, and workspace settings will be managed from this screen.',
    href: ROUTES.SETTINGS,
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
    href: ROUTES.DASHBOARD_METRICS,
    icon: BarChart3,
  },
  {
    slug: 'evaluations',
    label: 'Evaluations',
    title: 'Evaluations',
    detail: 'Evaluation results will appear here after benchmark runs complete.',
    href: ROUTES.DASHBOARD_EVALUATIONS,
    icon: GitBranch,
  },
  {
    slug: 'deployments',
    label: 'Deployments',
    title: 'Deployments',
    detail: 'Deployment status will appear here after environments publish releases.',
    href: ROUTES.DASHBOARD_DEPLOYMENTS,
    icon: Rocket,
  },
];

export const accountSection: DashboardSection = {
  slug: 'account',
  label: 'Account',
  title: 'Account',
  detail: 'Account identity and connected providers will appear here after sign-in.',
  href: ROUTES.LOGIN,
  icon: LockKeyhole,
};

export const routeSections = [...dashboardSections, ...dashboardTabs.filter((tab) => tab.slug !== 'overview')];

export function findDashboardSection(slug: string): DashboardSection | undefined {
  return routeSections.find((section) => section.slug === slug);
}

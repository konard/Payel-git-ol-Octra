export const ROUTES = {
  HOME: '/',
  WORKSPACE: '/app',
  LOGIN: '/login',
  LOGIN_GOOGLE: '/login/google',
  LOGIN_GITHUB: '/login/github',
  LOGIN_LEFINE: '/login/lefine',
  DASHBOARD: '/dashboard',
  DASHBOARD_FLOWS: '/dashboard/flows',
  DASHBOARD_MODELS: '/dashboard/models',
  DASHBOARD_FILES: '/dashboard/files',
  DASHBOARD_SECURITY: '/dashboard/security',
  DASHBOARD_METRICS: '/dashboard/metrics',
  DASHBOARD_EVALUATIONS: '/dashboard/evaluations',
  DASHBOARD_DEPLOYMENTS: '/dashboard/deployments',
  SETTINGS: '/settings',
} as const;

export const API = {
  LOGIN: '/login',
  REGISTER: '/register',
  ME: '/me',
} as const;

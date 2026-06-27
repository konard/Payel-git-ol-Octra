const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? '';

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
  PROFILE: '/profile',
  PROFILE_SETTINGS: '/profile/settings',
  PROFILE_API: '/profile/api',
} as const;

export const API = {
  BASE: API_BASE,
  LOGIN: `${API_BASE}/login`,
  REGISTER: `${API_BASE}/register`,
  ME: `${API_BASE}/me`,
} as const;

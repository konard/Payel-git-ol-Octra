import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();

function read(relativePath) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8');
}

function exists(relativePath) {
  return fs.existsSync(path.join(root, relativePath));
}

const failures = [];

function assert(condition, message) {
  if (!condition) failures.push(message);
}

const requiredFiles = [
  'next.config.mjs',
  'app/layout.tsx',
  'app/page.tsx',
  'app/login/page.tsx',
  'app/dashboard/page.tsx',
  'app/dashboard/[section]/page.tsx',
  'app/dashboard/DashboardShell.tsx',
  'app/dashboard/sections.ts',
  'app/components/EmptyDataPanel.tsx',
  'app/components/EmptyDataPanel.module.css',
  'app/components/UserBalance.tsx',
  'app/components/UserBalance.module.css',
  'app/globals.css',
];

for (const file of requiredFiles) {
  assert(exists(file), `Missing required Next frontend file: ${file}`);
}

if (exists('package.json')) {
  const packageJson = JSON.parse(read('package.json'));
  assert(packageJson.dependencies?.next, 'package.json must depend on next');
  assert(packageJson.dependencies?.['@xyflow/react'], 'package.json must depend on @xyflow/react for the requested React Flow canvas');
  assert(packageJson.scripts?.dev === 'next dev', 'dev script must run next dev');
  assert(packageJson.scripts?.build?.includes('next build'), 'build script must run next build');
  assert(!JSON.stringify(packageJson.scripts || {}).includes('vite'), 'package scripts must not invoke vite');
}

if (exists('next.config.mjs')) {
  const config = read('next.config.mjs');
  assert(config.includes("output: 'export'") || config.includes('output: "export"'), 'Next config must use static export output');
  assert(config.includes('unoptimized: true'), 'Next image optimization must be disabled for static export');
}

const allRouteSource = requiredFiles
  .filter((file) => file.endsWith('.tsx') && exists(file))
  .map(read)
  .join('\n')
  .toLowerCase();

for (const phrase of ['octra', 'google', 'github', 'lefine', 'dashboard', 'sign in', 'create account']) {
  assert(allRouteSource.includes(phrase), `New frontend source is missing required phrase: ${phrase}`);
}

if (exists('app/app/page.tsx')) {
  const homepageSource = [
    'app/app/page.tsx',
    'app/components/EnvironmentPanel.tsx',
    'app/lib/environments.ts',
    'app/config/routes.ts',
  ]
    .filter(exists)
    .map(read)
    .join('\n')
    .toLowerCase();
  for (const phrase of ['active-environments', 'environment_id', '/environment', '/api/chat', 'cli_state']) {
    assert(homepageSource.includes(phrase), `Homepage must present the requested TradingView-style node workspace: missing ${phrase}`);
  }

  assert(homepageSource.includes('userbalance'), 'Homepage top bar must render the server-backed UserBalance component');
  assert(homepageSource.includes('emptydatapanel'), 'Homepage runtime metrics must use an empty/live-data component instead of hardcoded metrics');

  for (const rejectedPhrase of ['const backendmetrics', '124k', '38m', 'proxy mode']) {
    assert(!homepageSource.includes(rejectedPhrase), `Homepage must not hardcode runtime metric data: ${rejectedPhrase}`);
  }

  for (const rejectedPhrase of ['ai delivery cockpit', 'hero-section', 'capability-band', 'live execution tape']) {
    assert(!homepageSource.includes(rejectedPhrase), `Homepage must not fall back to the rejected landing-page shape: ${rejectedPhrase}`);
  }
}

if (exists('app/page.tsx')) {
  const landingSource = read('app/page.tsx').toLowerCase();
  assert(
    landingSource.includes('emptydatapanel') || landingSource.includes('fake_metrics'),
    'Landing preview metrics must use shared empty-data UI or centralized preview data',
  );
  for (const rejectedPhrase of ['124k', '<strong>18</strong>', '<strong>7</strong>']) {
    assert(!landingSource.includes(rejectedPhrase), `Landing page must not hardcode preview metric data: ${rejectedPhrase}`);
  }
}

if (exists('app/dashboard/page.tsx')) {
  const dashboardSource = read('app/dashboard/page.tsx').toLowerCase();
  assert(dashboardSource.includes('dashboardshell'), 'Dashboard page must use the shared DashboardShell component');
  assert(dashboardSource.includes('emptydatapanel'), 'Dashboard page must use live-data empty states for missing metrics');
  for (const rejectedPhrase of ['processed prompts', '850ms', '$452.10', 'intent = billing', 'max cost']) {
    assert(!dashboardSource.includes(rejectedPhrase), `Dashboard page must not hardcode fake metrics or routing data: ${rejectedPhrase}`);
  }
}

if (exists('app/dashboard/[section]/page.tsx')) {
  const dashboardSectionSource = read('app/dashboard/[section]/page.tsx').toLowerCase();
  assert(dashboardSectionSource.includes('promise<{ section: string }>'), 'Dashboard section route must use Next 16 async params typing');
  assert(dashboardSectionSource.includes('await params'), 'Dashboard section route must await params before reading the section slug');
}

if (exists('app/dashboard/DashboardShell.tsx')) {
  const shellSource = read('app/dashboard/DashboardShell.tsx').toLowerCase();
  assert(shellSource.includes('userbalance'), 'Dashboard shell must expose the user balance in the top bar');
  assert(!shellSource.includes('href="#"'), 'Dashboard shell navigation must link to real screens, not placeholders');
}

if (exists('app/dashboard/sections.ts')) {
  const sectionsSource = read('app/dashboard/sections.ts').toLowerCase();
  for (const phrase of ['flows', 'models', 'files', 'security', 'settings', 'metrics', 'evaluations', 'deployments']) {
    assert(sectionsSource.includes(phrase), `Dashboard sections must include a screen for ${phrase}`);
  }
}

if (exists('app/components/UserBalance.tsx')) {
  const balanceSource = read('app/components/UserBalance.tsx').toLowerCase();
  const userServerSource = exists('app/server/user.ts') ? read('app/server/user.ts').toLowerCase() : '';
  const routeSource = exists('app/config/routes.ts') ? read('app/config/routes.ts').toLowerCase() : '';
  assert(
    balanceSource.includes('fetchme') && userServerSource.includes('api.me') && routeSource.includes('/me'),
    'UserBalance must load the authenticated user from /me',
  );
  assert(balanceSource.includes('balance'), 'UserBalance must display balance returned by the backend');
  assert(!balanceSource.includes('1250'), 'UserBalance must not hardcode a displayed credit balance');
}

if (exists('app/globals.css')) {
  const globalStyles = read('app/globals.css').toLowerCase();
  for (const phrase of ['--metric-success', '--metric-warning', '--metric-danger', '.metric-tone-success', '.metric-tone-warning', '.metric-tone-danger']) {
    assert(globalStyles.includes(phrase), `Metric styling must include requested color accents: missing ${phrase}`);
  }
}

const frontendWorkspaceSource = ['app/page.tsx', 'app/app/page.tsx']
  .filter(exists)
  .map(read)
  .join('\n')
  .toLowerCase();

for (const rejectedPhrase of ['boss', 'manager', 'worker']) {
  assert(!frontendWorkspaceSource.includes(rejectedPhrase), `Frontend workspace must not mention removed role hierarchy: ${rejectedPhrase}`);
}

if (exists('Dockerfile')) {
  const dockerfile = read('Dockerfile');
  assert(dockerfile.includes('/app/dist'), 'Dockerfile must continue serving the static dist directory');
}

if (failures.length > 0) {
  console.error('Next frontend check failed:');
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exit(1);
}

console.log('Next frontend structure, routes, auth providers, and static export settings are present.');

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
  'app/auth/page.tsx',
  'app/dashboard/page.tsx',
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
  const homepageSource = read('app/app/page.tsx').toLowerCase();
  for (const phrase of ['active-environments', 'environment_id', '/environment', '/api/chat', 'cli_state']) {
    assert(homepageSource.includes(phrase), `Homepage must present the requested TradingView-style node workspace: missing ${phrase}`);
  }

  for (const phrase of ['tone:', 'metric-tone-', 'runtime metrics']) {
    assert(homepageSource.includes(phrase), `Homepage runtime metrics must expose selective color tone data: missing ${phrase}`);
  }

  for (const rejectedPhrase of ['ai delivery cockpit', 'hero-section', 'capability-band', 'live execution tape']) {
    assert(!homepageSource.includes(rejectedPhrase), `Homepage must not fall back to the rejected landing-page shape: ${rejectedPhrase}`);
  }
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

for (const rejectedPhrase of ['boss', 'manager', 'worker', 'role']) {
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

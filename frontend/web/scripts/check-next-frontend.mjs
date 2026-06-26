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

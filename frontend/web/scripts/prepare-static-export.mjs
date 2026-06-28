import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const outDir = path.join(root, 'out');
const distDir = path.join(root, 'dist');

if (!fs.existsSync(outDir)) {
  console.error('Next static export directory is missing. Expected: out/');
  process.exit(1);
}

fs.rmSync(distDir, { recursive: true, force: true });
fs.cpSync(outDir, distDir, { recursive: true });

console.log('Copied Next static export from out/ to dist/.');

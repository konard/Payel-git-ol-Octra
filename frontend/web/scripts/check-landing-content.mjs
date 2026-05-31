import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();

const requiredFiles = [
  'src/app/components/LandingPage.tsx',
  'src/app/components/landing/HeroSection.tsx',
  'src/app/components/landing/WorkflowDemo.tsx',
  'src/app/components/landing/ProvidersSection.tsx',
  'src/app/components/landing/FooterSection.tsx',
];

const sourceText = requiredFiles
  .map((file) => fs.readFileSync(path.join(root, file), 'utf8'))
  .join('\n');

const requiredCopy = [
  'presentations',
  'text documents',
  'research',
  'developers',
  'regular users',
];

const missingCopy = requiredCopy.filter((phrase) => !sourceText.toLowerCase().includes(phrase));
if (missingCopy.length > 0) {
  console.error(`Landing page is missing required audience copy: ${missingCopy.join(', ')}`);
  process.exit(1);
}

const requiredAssets = [
  'src/images/main/landing/document-reader-dark.png',
  'src/images/main/landing/presentation-deck-dark.png',
  'src/images/main/landing/research-progress-dark.png',
  'src/images/main/landing/code-view-dark.png',
];

const missingAssets = requiredAssets.filter((file) => !fs.existsSync(path.join(root, file)));
if (missingAssets.length > 0) {
  console.error(`Landing page is missing refreshed screenshot assets: ${missingAssets.join(', ')}`);
  process.exit(1);
}

console.log('Landing content and refreshed screenshots are present.');

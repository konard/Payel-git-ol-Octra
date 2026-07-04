import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();

function read(relativePath) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8');
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

const modalSource = read('app/components/EditNodeModal.tsx');
const styles = read('app/globals.css');

assert(modalSource.includes("from './Select'"), 'Adapter protocol picker must use the shared custom Select component');
assert(!/<select[\s>]/.test(modalSource), 'Adapter protocol picker must not render a native select element');

for (const phrase of [
  "key: 'api_key'",
  "label: 'API Key'",
  'requiredFields',
  'environment_id',
  'apiKey',
  'x-api-key',
  'octra-api-token',
]) {
  assert(modalSource.includes(phrase), `Adapter modal must expose required contract field: ${phrase}`);
}

for (const phrase of [
  '.adapter-contract-fields',
  '.adapter-contract-field',
  '.adapter-contract-row pre',
]) {
  assert(styles.includes(phrase), `Adapter contract styling is missing ${phrase}`);
}

console.log('Adapter modal exposes styled protocol selection and required API-key contract fields.');

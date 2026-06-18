import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { dirname, resolve } from 'node:path';

// Feature test for issue #69: integrate with the ocawe framework
// (https://github.com/lefinepro/ocawe) for describing workflows. The Workflow
// Export modal must be able to render the current Boss → Manager → Worker graph
// as an ocawe Cawfile bundle in addition to the native JSON format.

const here = dirname(fileURLToPath(import.meta.url));
const root = resolve(here, '..');
const read = (rel) => readFileSync(resolve(root, rel), 'utf8');

// --- 1. The ocawe service exists and exposes the converter API -------------
const service = read('src/services/ocaweService.ts');
for (const symbol of [
  'export function buildOcaweBundle',
  'export function renderOcaweBundleText',
  'export function exportOcaweText',
  'export async function triggerOcaweWorkflow',
  'export function ocaweSlug',
]) {
  assert.ok(service.includes(symbol), `ocaweService.ts must export "${symbol}"`);
}
// The generated Cawfile must use the documented ocawe directives.
assert.match(service, /import = \[/, 'Cawfile must import agent markdown files');
assert.match(service, /\.\/agents\/\*\.md/, 'Cawfile must import ./agents/*.md');
assert.match(service, /workflow \$\{quote\(workflowName\)\} do/, 'Cawfile must declare a workflow block');
assert.match(service, /\/v1\/triggers\/workflows\//, 'trigger helper must call the ocawe runtime trigger API');

// --- 2. Behavioural check of the converter (transpiled via esbuild) --------
// esbuild ships as a Vite dependency, so it is available after `npm ci`.
let esbuild = null;
try {
  esbuild = await import('esbuild');
} catch {
  console.log('check-ocawe-export: esbuild unavailable, skipping behavioural assertions');
}

if (esbuild) {
  const { code } = await esbuild.transform(service, { loader: 'ts', format: 'esm' });
  const dataUrl = 'data:text/javascript;base64,' + Buffer.from(code).toString('base64');
  const mod = await import(dataUrl);

  const nodes = [
    { id: 'worker-1', type: 'worker', role: 'React Developer', techStack: ['react', 'vite'] },
    { id: 'boss-1', type: 'boss', role: 'CEO' },
    { id: 'manager-1', type: 'manager', role: 'Backend Manager' },
  ];
  const edges = [
    { from: 'boss-1', to: 'manager-1' },
    { from: 'manager-1', to: 'worker-1' },
  ];

  const bundle = mod.buildOcaweBundle(nodes, edges, { workflowName: 'My Flow' });
  assert.equal(bundle.workflowName, 'my-flow', 'workflow name must be slugified');
  assert.equal(bundle.agents.length, 3, 'one agent file per node');

  // Cawfile must be valid-looking ocawe and list agents Boss → Manager → Worker.
  assert.match(bundle.cawfile, /settings do/, 'Cawfile must contain a settings block');
  assert.match(bundle.cawfile, /workflow "my-flow" do/, 'Cawfile must declare the workflow');
  const ceoIdx = bundle.cawfile.indexOf('agent "ceo"');
  const mgrIdx = bundle.cawfile.indexOf('agent "backend-manager"');
  const devIdx = bundle.cawfile.indexOf('agent "react-developer"');
  assert.ok(ceoIdx >= 0 && mgrIdx >= 0 && devIdx >= 0, 'all agents must be referenced in the Cawfile');
  assert.ok(ceoIdx < mgrIdx && mgrIdx < devIdx, 'agents must be ordered Boss → Manager → Worker');

  // Each agent file is a markdown doc with frontmatter.
  const dev = bundle.agents.find((a) => a.path === 'agents/react-developer.md');
  assert.ok(dev, 'worker agent file must be named after its role');
  assert.match(dev.content, /^---\n/, 'agent file must start with frontmatter');
  assert.match(dev.content, /name: "React Developer"/, 'agent frontmatter must carry the role name');
  assert.match(dev.content, /react, vite/, 'agent file must mention the tech stack');

  // The single-document bundle must be split-able by file markers.
  const text = mod.exportOcaweText(nodes, edges, { workflowName: 'My Flow' });
  assert.match(text, /# ==== FILE: Cawfile ====/, 'bundle text must contain the Cawfile marker');
  assert.match(text, /# ==== FILE: agents\/ceo\.md ====/, 'bundle text must contain agent file markers');

  // Duplicate roles must get distinct file names.
  const dup = mod.buildOcaweBundle(
    [
      { id: 'w1', type: 'worker', role: 'Tester' },
      { id: 'w2', type: 'worker', role: 'Tester' },
    ],
    [],
  );
  const paths = dup.agents.map((a) => a.path);
  assert.equal(new Set(paths).size, paths.length, 'duplicate roles must produce unique agent files');

  // Empty graphs must still produce a valid Cawfile.
  const empty = mod.buildOcaweBundle([], []);
  assert.match(empty.cawfile, /workflow "octra-workflow" do/, 'empty graph must still yield a workflow');
}

// --- 3. The export modal wires up the ocawe format -------------------------
const modal = read('src/components/workspace/WorkflowExport.tsx');
assert.match(modal, /from '\.\.\/\.\.\/services\/ocaweService'/, 'modal must use the ocawe service');
assert.match(modal, /exportOcaweText/, 'modal must call exportOcaweText');
assert.match(modal, /workflowLibrary\.formatOcawe/, 'modal must render the Ocawe format toggle');
assert.match(modal, /workflowLibrary\.formatJson/, 'modal must render the JSON format toggle');
assert.match(modal, /handleFormatChange/, 'modal must switch between formats');

// --- 4. Every shipped translation defines the new keys ---------------------
const NEW_KEYS = ['formatLabel', 'formatJson', 'formatOcawe', 'ocaweHint', 'copyBundle', 'downloadBundle'];
for (const dir of ['languages', 'public/languages']) {
  for (const file of readdirSync(resolve(root, dir)).filter((f) => f.endsWith('.json'))) {
    const data = JSON.parse(read(`${dir}/${file}`));
    if (!data.workflowLibrary) continue; // some locales may omit the section entirely
    for (const key of NEW_KEYS) {
      assert.ok(
        typeof data.workflowLibrary[key] === 'string' && data.workflowLibrary[key].length > 0,
        `${dir}/${file} must define workflowLibrary.${key}`,
      );
    }
  }
}

console.log('check-ocawe-export: all assertions passed');

/**
 * Ocawe Integration Service
 *
 * Ocawe (https://github.com/lefinepro/ocawe) is a Crystal-first runtime for
 * workflow bundles, agents, tools and skills. Workflows are described as
 * "Cawfile" bundles (RCL format) that import one markdown file per agent.
 *
 * This service converts an Octra Boss → Manager → Worker graph into an ocawe
 * workflow bundle so a workflow designed visually in Octra can be described and
 * executed by the ocawe runtime. It also exposes a trigger helper that calls the
 * ocawe runtime HTTP API (`POST /v1/triggers/workflows/:id`).
 *
 * The conversion is intentionally pure and deterministic (no timestamps / random
 * ids) so the output is stable and unit-testable.
 */

export interface OcaweNode {
  id: string;
  type: string; // 'boss' | 'manager' | 'worker' | 'universal' | 'github' | ...
  role?: string;
  techStack?: string[];
  [key: string]: unknown;
}

export interface OcaweEdge {
  from: string;
  to: string;
}

export interface OcaweExportOptions {
  /** Workflow id used in the Cawfile (`workflow "<name>" do`). */
  workflowName?: string;
  /** Default model directive, e.g. "cliproxy". */
  model?: string;
  /** Runtime port written into the `settings` block. */
  port?: number;
}

export interface OcaweFile {
  /** Path relative to the bundle root, e.g. "agents/boss-1.md". */
  path: string;
  content: string;
}

export interface OcaweBundle {
  /** Workflow id used inside the bundle. */
  workflowName: string;
  /** The Cawfile describing the workflow. */
  cawfile: string;
  /** One markdown agent definition per node. */
  agents: OcaweFile[];
}

const DEFAULT_MODEL = 'cliproxy';
const DEFAULT_PORT = 4111;

/**
 * Slugify an arbitrary string into a safe ocawe identifier / file name.
 * Falls back to a deterministic value so empty input never yields "".
 */
export function ocaweSlug(value: string, fallback = 'agent'): string {
  const slug = (value || '')
    .toString()
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
  return slug || fallback;
}

/** Escape a value for use inside an RCL/Crystal double-quoted string. */
function quote(value: string): string {
  return `"${String(value).replace(/\\/g, '\\\\').replace(/"/g, '\\"')}"`;
}

/**
 * Order nodes so the Cawfile reads top-down: Boss first, then Managers, then
 * Workers. Within the graph we honour edges (a node never appears before the
 * node that points at it) and fall back to a stable type-based ordering.
 */
function orderNodes(nodes: OcaweNode[], edges: OcaweEdge[]): OcaweNode[] {
  const typeRank: Record<string, number> = { boss: 0, manager: 1, worker: 2, universal: 2, github: 3 };
  const rankOf = (n: OcaweNode) => (n.type in typeRank ? typeRank[n.type] : 9);

  // Incoming-edge count for a light topological bias.
  const incoming = new Map<string, number>();
  nodes.forEach((n) => incoming.set(n.id, 0));
  edges.forEach((e) => {
    if (incoming.has(e.to)) incoming.set(e.to, (incoming.get(e.to) || 0) + 1);
  });

  return [...nodes].sort((a, b) => {
    const ra = rankOf(a);
    const rb = rankOf(b);
    if (ra !== rb) return ra - rb;
    const ia = incoming.get(a.id) || 0;
    const ib = incoming.get(b.id) || 0;
    if (ia !== ib) return ia - ib;
    return String(a.id).localeCompare(String(b.id));
  });
}

/** A short human description for an agent derived from its node type/role. */
function describeNode(node: OcaweNode): string {
  const role = node.role || node.type;
  switch (node.type) {
    case 'boss':
      return `Boss agent (${role}) — plans the architecture and coordinates managers.`;
    case 'manager':
      return `Manager agent (${role}) — reviews and orchestrates workers.`;
    case 'worker':
      return `Worker agent (${role}) — implements the assigned task.`;
    case 'universal':
      return `Universal agent (${role}) — solves easy tasks directly without a full team.`;
    case 'github':
      return `GitHub agent (${role}) — handles repository operations.`;
    default:
      return `${role} agent.`;
  }
}

/** Build a single agent markdown file (frontmatter + system prompt). */
function buildAgentFile(node: OcaweNode, model: string): OcaweFile {
  const slug = ocaweSlug(node.role || node.type || node.id, ocaweSlug(node.id));
  const name = node.role || node.type || node.id;
  const description = describeNode(node);
  const stack =
    Array.isArray(node.techStack) && node.techStack.length > 0
      ? node.techStack.join(', ')
      : '';

  const lines = [
    '---',
    `name: ${quote(name)}`,
    `description: ${quote(description)}`,
    `model: ${quote(model)}`,
    '---',
    '',
    describeNode(node),
  ];
  if (stack) {
    lines.push('', `Preferred tech stack: ${stack}.`);
  }
  lines.push('');

  return { path: `agents/${slug}.md`, content: lines.join('\n') };
}

/**
 * Convert an Octra graph into a complete ocawe workflow bundle.
 */
export function buildOcaweBundle(
  nodes: OcaweNode[],
  edges: OcaweEdge[] = [],
  options: OcaweExportOptions = {},
): OcaweBundle {
  const workflowName = ocaweSlug(options.workflowName || 'octra-workflow', 'octra-workflow');
  const model = options.model || DEFAULT_MODEL;
  const port = options.port || DEFAULT_PORT;

  const ordered = orderNodes(nodes || [], edges || []);

  // Deduplicate agent slugs so two nodes with the same role get distinct files.
  const usedSlugs = new Set<string>();
  const agentFiles: OcaweFile[] = [];
  const agentRefs: string[] = [];

  ordered.forEach((node) => {
    const file = buildAgentFile(node, model);
    let path = file.path;
    let slug = path.replace(/^agents\//, '').replace(/\.md$/, '');
    let suffix = 2;
    while (usedSlugs.has(slug)) {
      slug = `${path.replace(/^agents\//, '').replace(/\.md$/, '')}-${suffix++}`;
      path = `agents/${slug}.md`;
    }
    usedSlugs.add(slug);
    agentFiles.push({ path, content: file.content });
    agentRefs.push(slug);
  });

  const agentSteps =
    agentRefs.length > 0
      ? agentRefs.map((ref) => `  agent ${quote(ref)}`).join('\n')
      : '  # No agents defined — add agents to the workflow canvas first.';

  const cawfile = [
    `# Ocawe workflow bundle generated by Octra`,
    `# Run with: ocawe up`,
    '',
    'settings do',
    '  data.adapter = "memory"',
    `  port = ${port}`,
    'end',
    '',
    'import = [',
    '  "./agents/*.md"',
    ']',
    '',
    `@[Model(${quote(model)})]`,
    `workflow ${quote(workflowName)} do`,
    agentSteps,
    'end',
    '',
  ].join('\n');

  return { workflowName, cawfile, agents: agentFiles };
}

/**
 * Render an ocawe bundle as a single self-contained text document so it can be
 * previewed in a textarea and downloaded as one file. Files are separated by
 * `# ==== FILE: <path> ====` markers so the bundle can be split back out.
 */
export function renderOcaweBundleText(bundle: OcaweBundle): string {
  const parts: string[] = [];
  parts.push(`# ==== FILE: Cawfile ====`);
  parts.push(bundle.cawfile.replace(/\n+$/, ''));
  for (const agent of bundle.agents) {
    parts.push('');
    parts.push(`# ==== FILE: ${agent.path} ====`);
    parts.push(agent.content.replace(/\n+$/, ''));
  }
  return parts.join('\n') + '\n';
}

/** Convenience: graph → single ocawe bundle text document. */
export function exportOcaweText(
  nodes: OcaweNode[],
  edges: OcaweEdge[] = [],
  options: OcaweExportOptions = {},
): string {
  return renderOcaweBundleText(buildOcaweBundle(nodes, edges, options));
}

/**
 * Trigger a workflow on a running ocawe runtime.
 * Calls `POST /v1/triggers/workflows/:id` as documented in the ocawe README.
 */
export async function triggerOcaweWorkflow(
  baseUrl: string,
  workflowId: string,
  input: Record<string, unknown> = {},
): Promise<unknown> {
  const url = `${baseUrl.replace(/\/$/, '')}/v1/triggers/workflows/${encodeURIComponent(workflowId)}`;
  const response = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ input }),
  });
  if (!response.ok) {
    throw new Error(`Failed to trigger ocawe workflow: ${response.status} ${response.statusText}`);
  }
  return response.json().catch(() => ({}));
}

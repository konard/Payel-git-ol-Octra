'use client';

import { Zap, Workflow, Bot, Puzzle, ExternalLink, Copy } from 'lucide-react';
import { DashboardShell } from '../DashboardShell';
import { API } from '../../config/routes';
import { useState } from 'react';

type TriggerEndpoint = {
  method: string;
  path: string;
  icon: typeof Zap;
  title: string;
  description: string;
  requestBody: string;
  responseExample: string;
};

const endpoints: TriggerEndpoint[] = [
  {
    method: 'POST',
    path: '/v1/triggers/workflows/:id',
    icon: Workflow,
    title: 'Trigger Workflow',
    description: 'Start a new workflow run by workflow ID. The request body is passed as input data to the workflow.',
    requestBody: JSON.stringify({ input_data: { key: 'value' } }, null, 2),
    responseExample: JSON.stringify({ workflow_id: '...', run_id: '...', status: 'running' }, null, 2),
  },
  {
    method: 'POST',
    path: '/v1/triggers/agents/:id',
    icon: Bot,
    title: 'Trigger Agent',
    description: 'Send a chat message to an AI agent by agent ID. The agent generates a response using its configured model and prompt.',
    requestBody: JSON.stringify({ messages: [{ role: 'user', content: 'Hello!' }] }, null, 2),
    responseExample: JSON.stringify({ id: 'trg_agent_...', object: 'trigger.agent.response', agent_id: '...', output_text: '...' }, null, 2),
  },
  {
    method: 'POST',
    path: '/v1/triggers/skills/:id',
    icon: Puzzle,
    title: 'Trigger Skill',
    description: 'Execute a skill by skill ID. The request body is passed as input to the skill.',
    requestBody: JSON.stringify({ input: { param1: 'value' } }, null, 2),
    responseExample: JSON.stringify({ id: 'trg_skill_...', object: 'trigger.skill.response', status: 'ok' }, null, 2),
  },
  {
    method: 'POST',
    path: '/v1/triggers/functions/:id',
    icon: Zap,
    title: 'Trigger Function',
    description: 'Call a registered function by function ID. The input is passed directly to the function handler.',
    requestBody: JSON.stringify({ input: { arg1: 'value' } }, null, 2),
    responseExample: JSON.stringify({ id: 'trg_fn_...', object: 'trigger.function.response', output: { result: '...' } }, null, 2),
  },
];

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <button
      className="icon-button"
      onClick={() => { navigator.clipboard.writeText(text).then(() => { setCopied(true); setTimeout(() => setCopied(false), 2000); }); }}
      aria-label="Copy"
      style={{ color: copied ? 'var(--success)' : 'var(--muted)' }}
    >
      <Copy size={14} />
    </button>
  );
}

export default function TriggersPage() {
  return (
    <DashboardShell activeSection="triggers" hideTabs showNotifications={false}>
      <section className="dashboard-grid dashboard-grid-single">
        <article className="large-panel" aria-label="Triggers & Webhooks" style={{ padding: 24, minHeight: 0 }}>
          <div className="panel-heading" style={{ padding: 0, marginBottom: 24, border: 'none' }}>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <Zap size={18} />
              <span>Triggers & Webhooks</span>
            </div>
            <p style={{ margin: '8px 0 0', color: 'var(--muted)', fontSize: '0.85rem', fontWeight: 400 }}>
              HTTP endpoints for automating workflows, agents, skills, and functions. Use these from any HTTP client.
            </p>
          </div>

          <div style={{ display: 'grid', gap: 16 }}>
            {endpoints.map((ep) => {
              const Icon = ep.icon;
              const fullUrl = `${API.BASE || 'https://octra.ai'}${ep.path}`;
              return (
                <div key={ep.path} style={{ padding: 16, border: '1px solid var(--line)', borderRadius: 8, background: 'var(--surface)' }}>
                  <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12, marginBottom: 12 }}>
                    <div style={{ padding: 8, background: 'var(--bg)', borderRadius: 8, color: 'var(--accent)', flexShrink: 0 }}>
                      <Icon size={20} />
                    </div>
                    <div style={{ flex: 1 }}>
                      <h3 style={{ margin: '0 0 4px', fontSize: '0.95rem' }}>{ep.title}</h3>
                      <p style={{ margin: '0 0 8px', color: 'var(--muted)', fontSize: '0.82rem' }}>{ep.description}</p>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                        <span style={{ padding: '2px 6px', borderRadius: 4, background: 'var(--accent)', color: '#fff', fontSize: '0.72rem', fontWeight: 600, fontFamily: 'monospace' }}>{ep.method}</span>
                        <code style={{ fontSize: '0.82rem', color: 'var(--text)', wordBreak: 'break-all' }}>{fullUrl}</code>
                        <CopyButton text={fullUrl} />
                      </div>
                    </div>
                  </div>

                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                    <div>
                      <h4 style={{ margin: '0 0 4px', fontSize: '0.78rem', color: 'var(--quiet)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Request Body</h4>
                      <pre style={{ margin: 0, padding: 10, background: 'var(--bg)', borderRadius: 6, fontSize: '0.75rem', overflow: 'auto', maxHeight: 120, border: '1px solid var(--line)' }}>{ep.requestBody}</pre>
                    </div>
                    <div>
                      <h4 style={{ margin: '0 0 4px', fontSize: '0.78rem', color: 'var(--quiet)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Response</h4>
                      <pre style={{ margin: 0, padding: 10, background: 'var(--bg)', borderRadius: 6, fontSize: '0.75rem', overflow: 'auto', maxHeight: 120, border: '1px solid var(--line)' }}>{ep.responseExample}</pre>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          <div style={{ marginTop: 24, padding: 16, border: '1px dashed var(--line)', borderRadius: 8, background: 'var(--surface)' }}>
            <h3 style={{ margin: '0 0 8px', fontSize: '0.9rem', display: 'flex', alignItems: 'center', gap: 6 }}>
              <ExternalLink size={16} /> Using Triggers
            </h3>
            <p style={{ margin: '0 0 8px', color: 'var(--muted)', fontSize: '0.82rem' }}>
              All trigger endpoints require authentication via the <code style={{ fontSize: '0.78rem' }}>octra-api-token</code> header.
              Replace <code style={{ fontSize: '0.78rem' }}>:id</code> with the actual workflow, agent, skill, or function ID.
            </p>
            <p style={{ margin: 0, color: 'var(--muted)', fontSize: '0.82rem' }}>
              Example using curl:
            </p>
            <pre style={{ margin: '8px 0 0', padding: 10, background: 'var(--bg)', borderRadius: 6, fontSize: '0.78rem', overflow: 'auto', border: '1px solid var(--line)' }}>
{`curl -X POST ${API.BASE || 'https://octra.ai'}/v1/triggers/workflows/YOUR_WORKFLOW_ID \\
  -H "octra-api-token: YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"input_data": {"message": "hello"}}'`}
            </pre>
          </div>
        </article>
      </section>
    </DashboardShell>
  );
}

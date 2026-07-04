'use client';

import { Copy, Check, X } from 'lucide-react';
import { useState, useEffect, useCallback } from 'react';
import type { WorkflowCanvasItem } from './WorkflowCanvas';

type EditNodeModalProps = {
  item: WorkflowCanvasItem;
  onSave: (updated: WorkflowCanvasItem) => void;
  onClose: () => void;
};

type FieldDef = {
  key: string;
  label: string;
  type: 'text' | 'password' | 'select';
  options?: { value: string; label: string }[];
};

function fieldsForKind(kind: WorkflowCanvasItem['kind']): FieldDef[] {
  switch (kind) {
    case 'provider':
    case 'custom_provider':
      return [
        { key: 'auth', label: 'API Key', type: 'password' },
        { key: 'base_url', label: 'Base URL', type: 'text' },
        { key: 'model', label: 'Model', type: 'text' },
      ];
    case 'cli':
      return [
        { key: 'cli', label: 'Install method', type: 'text' },
      ];
    case 'skill':
      return [
        { key: 'skill', label: 'Skill name / ID', type: 'text' },
        { key: 'cli', label: 'Install method', type: 'text' },
      ];
    case 'adapter':
      return [
        {
          key: 'protocol',
          label: 'Protocol',
          type: 'select',
          options: [
            { value: 'grpc', label: 'gRPC' },
            { value: 'graphql', label: 'GraphQL' },
            { value: 'websocket', label: 'WebSocket' },
          ],
        },
        { key: 'path', label: 'Path', type: 'text' },
      ];
    default:
      return [];
  }
}

const adapterContractSchemas: Record<string, { endpoint: string; method: string; headers: string; body: string }> = {
  grpc: {
    endpoint: 'grpc://<host>:9090/chat.ChatService/Chat',
    method: 'Bidirectional streaming RPC',
    headers: 'x-api-key: <api_key>\nx-environment-id: <environment_id>',
    body: `{
  "prompt": "Hello, agent!"
}`,
  },
  graphql: {
    endpoint: 'http://<host>:8080/graphql',
    method: 'POST',
    headers: `Content-Type: application/json`,
    body: `{
  "query": "mutation { chat(environmentId: \\\"<env_id>\\\", prompt: \\\"Hello\\\", apiKey: \\\"<api_key>\\\") }"
}`,
  },
  websocket: {
    endpoint: 'ws://<host>:8080/ws/chat/<environment_id>',
    method: 'WebSocket (bidirectional)',
    headers: 'x-api-key: <api_key> (query param: ?token=<api_key>)',
    body: `{
  "prompt": "Hello, agent!"
}`,
  },
};

export function EditNodeModal({ item, onSave, onClose }: EditNodeModalProps) {
  const [name, setName] = useState(item.name);
  const [copied, setCopied] = useState(false);
  const [meta, setMeta] = useState<Record<string, string>>(() => {
    const m: Record<string, string> = {};
    for (const [k, v] of Object.entries(item.meta ?? {})) {
      if (v !== undefined) m[k] = v;
    }
    return m;
  });

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [onClose]);

  const fields = fieldsForKind(item.kind);
  const isAdapter = item.kind === 'adapter';
  const protocol = meta.protocol || 'websocket';
  const contract = isAdapter ? adapterContractSchemas[protocol] : null;

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    onSave({
      ...item,
      name,
      meta: { ...item.meta, ...meta },
    });
  }

  const handleCopyReference = useCallback(async () => {
    if (!contract) return;
    const ref = `# ${protocol.toUpperCase()} Adapter\n\nEndpoint: ${contract.endpoint}\nMethod: ${contract.method}\nHeaders:\n${contract.headers}\n\nBody:\n${contract.body}`;
    await navigator.clipboard.writeText(ref);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [contract, protocol]);

  function updateMeta(key: string, value: string) {
    setMeta((prev) => ({ ...prev, [key]: value }));
  }

  return (
    <div className="overlay" onClick={onClose} role="presentation">
      <div className="edit-node-modal" onClick={(e) => e.stopPropagation()} role="dialog" aria-label="Edit node">
        <div className="edit-node-header">
          <h2>Edit {nodeKindLabel(item.kind)}</h2>
          <button className="icon-button" onClick={onClose} aria-label="Close">
            <X size={15} />
          </button>
        </div>
        <form onSubmit={handleSubmit}>
          <label className="edit-node-field">
            <span>Name</span>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Node name"
            />
          </label>
          {fields.map((f) => (
            <label className="edit-node-field" key={f.key}>
              <span>{f.label}</span>
              {f.type === 'select' && f.options ? (
                <div className="custom-select" style={{ width: '100%' }}>
                  <select
                    className="edit-node-select"
                    value={meta[f.key] ?? f.options[0].value}
                    onChange={(e) => updateMeta(f.key, e.target.value)}
                  >
                    {f.options.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                </div>
              ) : (
                <input
                  type={f.type}
                  value={meta[f.key] ?? ''}
                  onChange={(e) => setMeta((prev) => ({ ...prev, [f.key]: e.target.value }))}
                  placeholder={f.label}
                />
              )}
            </label>
          ))}

          {isAdapter && contract && (
            <div className="adapter-contract">
              <div className="adapter-contract-header">
                <span>Contract schema — {protocol.toUpperCase()}</span>
                <button
                  type="button"
                  className="icon-button"
                  onClick={handleCopyReference}
                  aria-label="Copy reference"
                  title="Copy reference"
                >
                  {copied ? <Check size={15} /> : <Copy size={15} />}
                </button>
              </div>
              <div className="adapter-contract-block">
                <div className="adapter-contract-row">
                  <span className="adapter-contract-label">Endpoint</span>
                  <code>{contract.endpoint}</code>
                </div>
                <div className="adapter-contract-row">
                  <span className="adapter-contract-label">Method</span>
                  <code>{contract.method}</code>
                </div>
                <div className="adapter-contract-row">
                  <span className="adapter-contract-label">Headers</span>
                  <pre>{contract.headers}</pre>
                </div>
                <div className="adapter-contract-row">
                  <span className="adapter-contract-label">Body</span>
                  <pre>{contract.body}</pre>
                </div>
              </div>
            </div>
          )}

          <div className="edit-node-actions">
            <button type="button" className="secondary-button" onClick={onClose}>
              Cancel
            </button>
            <button className="primary-button" type="submit">
              Save
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function nodeKindLabel(kind: WorkflowCanvasItem['kind']) {
  switch (kind) {
    case 'provider':
      return 'Provider';
    case 'cli':
      return 'CLI';
    case 'skill':
      return 'Skill';
    case 'custom_provider':
      return 'Custom Provider';
    case 'environment':
      return 'Environment';
    case 'adapter':
      return 'Adapter';
  }
}

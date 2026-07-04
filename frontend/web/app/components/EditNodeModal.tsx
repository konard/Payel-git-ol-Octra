'use client';

import { Copy, Check, X } from 'lucide-react';
import { useState, useEffect, useCallback } from 'react';
import type { WorkflowCanvasItem } from './WorkflowCanvas';
import { Select } from './Select';

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
  placeholder?: string;
  required?: boolean;
};

const adapterProtocolOptions = [
  { value: 'grpc', label: 'gRPC' },
  { value: 'graphql', label: 'GraphQL' },
  { value: 'websocket', label: 'WebSocket' },
];

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
          options: adapterProtocolOptions,
          required: true,
        },
        { key: 'path', label: 'Path', type: 'text', required: true },
        { key: 'api_key', label: 'API Key', type: 'password', placeholder: '<api_key>', required: true },
      ];
    default:
      return [];
  }
}

type AdapterContractField = {
  label: string;
  key: string;
  description: string;
};

type AdapterContractSchema = {
  path: string;
  endpoint: string;
  method: string;
  requiredFields: AdapterContractField[];
  headers: string;
  body: string;
};

const adapterContractSchemas: Record<string, AdapterContractSchema> = {
  grpc: {
    path: 'chat.ChatService/Chat',
    endpoint: 'grpc://<host>:9090/chat.ChatService/Chat',
    method: 'Bidirectional streaming RPC',
    requiredFields: [
      { label: 'API Key', key: 'x-api-key', description: 'User API key passed in metadata.' },
      { label: 'Environment ID', key: 'x-environment-id', description: 'Dashboard environment UUID.' },
      { label: 'Prompt', key: 'prompt', description: 'ChatRequest prompt text.' },
    ],
    headers: 'x-api-key: <api_key>\nx-environment-id: <environment_id>',
    body: `ChatRequest {
  prompt: "Hello, agent!"
}`,
  },
  graphql: {
    path: '/graphql',
    endpoint: 'http://<host>:8080/graphql',
    method: 'POST',
    requiredFields: [
      { label: 'API Key', key: 'apiKey', description: 'Required chat mutation argument.' },
      { label: 'Environment ID', key: 'environmentId', description: 'Dashboard environment UUID.' },
      { label: 'Prompt', key: 'prompt', description: 'Message sent to the environment.' },
    ],
    headers: `Content-Type: application/json`,
    body: `{
  "query": "mutation { chat(environmentId: \\\"<environment_id>\\\", prompt: \\\"Hello, agent!\\\", apiKey: \\\"<api_key>\\\") }"
}`,
  },
  websocket: {
    path: '/ws/chat/<environment_id>',
    endpoint: 'ws://<host>:8080/ws/chat/<environment_id>?token=<api_key>',
    method: 'WebSocket (bidirectional)',
    requiredFields: [
      { label: 'API Key', key: 'token', description: 'Query parameter, or octra-api-token header.' },
      { label: 'Environment ID', key: 'environment_id', description: 'Environment UUID in the URL path.' },
      { label: 'Prompt', key: 'prompt', description: 'JSON message field sent over the socket.' },
    ],
    headers: 'octra-api-token: <api_key> (optional when ?token=<api_key> is used)',
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
    if (item.kind === 'adapter') {
      const protocol = m.protocol || 'websocket';
      m.protocol = protocol;
      if (!m.path) m.path = adapterContractSchemas[protocol]?.path ?? adapterContractSchemas.websocket.path;
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
  const contract = isAdapter ? (adapterContractSchemas[protocol] ?? adapterContractSchemas.websocket) : null;

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
    const required = contract.requiredFields
      .map((field) => `- ${field.label} (${field.key}): ${field.description}`)
      .join('\n');
    const ref = `# ${protocol.toUpperCase()} Adapter\n\nEndpoint: ${contract.endpoint}\nMethod: ${contract.method}\nRequired fields:\n${required}\n\nHeaders:\n${contract.headers}\n\nBody:\n${contract.body}`;
    await navigator.clipboard.writeText(ref);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [contract, protocol]);

  function updateMeta(key: string, value: string) {
    setMeta((prev) => {
      if (key !== 'protocol') return { ...prev, [key]: value };

      const currentProtocol = prev.protocol || 'websocket';
      const currentDefaultPath = adapterContractSchemas[currentProtocol]?.path;
      const nextDefaultPath = adapterContractSchemas[value]?.path;
      const shouldUpdatePath = !prev.path || prev.path === currentDefaultPath;

      return {
        ...prev,
        protocol: value,
        path: shouldUpdatePath && nextDefaultPath ? nextDefaultPath : prev.path,
      };
    });
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
            <div className="edit-node-field" key={f.key}>
              <span>
                {f.label}
                {f.required ? <b className="edit-node-required">Required</b> : null}
              </span>
              {f.type === 'select' && f.options ? (
                <Select
                  options={f.options}
                  value={meta[f.key] ?? f.options[0].value}
                  onChange={(value) => updateMeta(f.key, value)}
                  ariaLabel={f.label}
                  className="edit-node-protocol-select"
                />
              ) : (
                <input
                  type={f.type}
                  value={meta[f.key] ?? ''}
                  onChange={(e) => updateMeta(f.key, e.target.value)}
                  placeholder={f.placeholder ?? f.label}
                  required={f.required}
                  autoComplete={f.type === 'password' ? 'off' : undefined}
                />
              )}
            </div>
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
              <div className="adapter-contract-fields" aria-label="Required contract fields">
                {contract.requiredFields.map((field) => (
                  <div className="adapter-contract-field" key={field.key}>
                    <span>{field.label}</span>
                    <code>{field.key}</code>
                    <small>{field.description}</small>
                  </div>
                ))}
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

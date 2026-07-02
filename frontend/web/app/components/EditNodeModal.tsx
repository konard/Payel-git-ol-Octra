'use client';

import { X } from 'lucide-react';
import { useState, useEffect } from 'react';
import type { WorkflowCanvasItem } from './WorkflowCanvas';

type EditNodeModalProps = {
  item: WorkflowCanvasItem;
  onSave: (updated: WorkflowCanvasItem) => void;
  onClose: () => void;
};

type FieldDef = {
  key: string;
  label: string;
  type: 'text' | 'password';
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
    default:
      return [];
  }
}

export function EditNodeModal({ item, onSave, onClose }: EditNodeModalProps) {
  const [name, setName] = useState(item.name);
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

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    onSave({
      ...item,
      name,
      meta: { ...item.meta, ...meta },
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
            <label className="edit-node-field" key={f.key}>
              <span>{f.label}</span>
              <input
                type={f.type}
                value={meta[f.key] ?? ''}
                onChange={(e) => setMeta((prev) => ({ ...prev, [f.key]: e.target.value }))}
                placeholder={f.label}
              />
            </label>
          ))}
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
  }
}

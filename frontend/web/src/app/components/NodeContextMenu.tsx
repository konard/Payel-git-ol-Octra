import { useEffect, useRef } from 'react';
import { useTaskStore, type AgentNodeType } from '../../stores/taskStore';
import { useI18n } from '../../hooks/useI18n';
import bossImage from '../../images/boss-image.png';
import managerImage from '../../images/manager-image.png';
import workerImage from '../../images/worker-image.png';

const nodeImages: Record<AgentNodeType, string> = {
  boss: bossImage,
  manager: managerImage,
  worker: workerImage,
  zip: bossImage,
};

interface ContextMenuProps {
  x: number;
  y: number;
  nodeId: string;
  nodeType: AgentNodeType;
  nodeRole: string;
  onClose: () => void;
}

const roleOptions: Record<AgentNodeType, string[]> = {
  boss: ['CEO', 'CTO', 'Technical Director', 'Architect'],
  manager: ['Backend', 'Frontend', 'Fullstack', 'DevOps', 'Mobile', 'QA Lead'],
  worker: ['Developer', 'UI Developer', 'API Developer', 'Database', 'Tester', 'Documentation'],
  zip: ['Archive'],
};

export function NodeContextMenu({ x, y, nodeId, nodeType, nodeRole, onClose }: ContextMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null);
  const updateNode = useTaskStore((state) => state.updateNode);
  const removeNode = useTaskStore((state) => state.removeNode);
  const { t } = useI18n();

  // Закрытие при клике вне меню
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onClose();
      }
    };

    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };

    document.addEventListener('mousedown', handleClick);
    document.addEventListener('keydown', handleEsc);
    return () => {
      document.removeEventListener('mousedown', handleClick);
      document.removeEventListener('keydown', handleEsc);
    };
  }, [onClose]);

  const handleRoleChange = (newRole: string) => {
    updateNode(nodeId, { role: newRole });
    onClose();
  };

  const handleDelete = () => {
    removeNode(nodeId);
    onClose();
  };

  const handleDuplicate = () => {
    const addNode = useTaskStore.getState().addNode;
    const currentNode = useTaskStore.getState().nodes.find((n) => n.id === nodeId);
    if (currentNode) {
      addNode({
        id: `node-${Date.now()}`,
        type: currentNode.type,
        role: currentNode.role,
        status: currentNode.status,
        progress: 0,
        position: {
          x: (currentNode.position?.x || 0) + 50,
          y: (currentNode.position?.y || 0) + 50,
        },
      });
    }
    onClose();
  };

  return (
    <div
      ref={menuRef}
      className="fixed z-[100] bg-[var(--surface)] border border-[var(--border)] rounded-lg shadow-xl min-w-[220px] overflow-hidden"
      style={{ left: x, top: y }}
    >
      {/* Header с изображением */}
      <div className="px-4 py-3 border-b border-[var(--border)] bg-[var(--background)] flex items-center gap-3">
        <img
          src={nodeImages[nodeType]}
          alt={nodeType}
          className="w-10 h-10 object-contain"
        />
        <div className="text-sm font-semibold text-[var(--text)]">
          {nodeType.toUpperCase()} — {nodeRole}
        </div>
      </div>

      {/* Смена роли */}
      <div className="p-2 border-b border-[var(--border)]">
        <div className="text-xs text-[var(--text-muted)] mb-1 px-2">{t('contextMenu.changeRole')}</div>
        {roleOptions[nodeType]?.map((role) => (
          <button
            key={role}
            onClick={() => handleRoleChange(role)}
            className={`w-full text-left px-3 py-1.5 text-sm rounded transition-colors ${
              role === nodeRole
                ? 'bg-[var(--accent)] text-white'
                : 'hover:bg-[var(--background)] text-[var(--text)]'
            }`}
          >
            {role}
          </button>
        ))}
      </div>

      {/* Действия */}
      <div className="p-2">
        <button
          onClick={handleDuplicate}
          className="w-full text-left px-3 py-1.5 text-sm rounded hover:bg-[var(--background)] text-[var(--text)] transition-colors flex items-center gap-2"
        >
          <span>Duplicate</span> {t('contextMenu.duplicate')}
        </button>
        <button
          onClick={handleDelete}
          className="w-full text-left px-3 py-1.5 text-sm rounded hover:bg-red-500/20 text-red-500 transition-colors flex items-center gap-2"
        >
          <span>Delete</span> {t('contextMenu.delete')}
        </button>
      </div>
    </div>
  );
}

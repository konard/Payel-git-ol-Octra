import { useState } from 'react';
import { useI18n } from '../../hooks/useI18n';
import bossImage from '../../images/boss-image.png';
import managerImage from '../../images/manager-image.png';
import workerImage from '../../images/worker-image.png';

interface NodeTemplate {
  type: 'boss' | 'manager' | 'worker';
  label: string;
  image: string;
  typeLabel: string;
  description: string;
  color: string;
}

interface NodeSidebarProps {
  isOpen: boolean;
  onClose: () => void;
  onDragStart: (type: 'boss' | 'manager' | 'worker', event: React.DragEvent) => void;
}

export function NodeSidebar({ isOpen, onClose, onDragStart }: NodeSidebarProps) {
  const { t } = useI18n();
  const [draggingType, setDraggingType] = useState<'boss' | 'manager' | 'worker' | null>(null);

  const templates: NodeTemplate[] = [
    {
      type: 'boss',
      label: t('sidebar.boss.label'),
      image: bossImage,
      typeLabel: t('sidebar.boss.type'),
      description: t('sidebar.boss.description'),
      color: 'from-orange-500 to-orange-600',
    },
    {
      type: 'manager',
      label: t('sidebar.manager.label'),
      image: managerImage,
      typeLabel: t('sidebar.manager.type'),
      description: t('sidebar.manager.description'),
      color: 'from-blue-500 to-blue-600',
    },
    {
      type: 'worker',
      label: t('sidebar.worker.label'),
      image: workerImage,
      typeLabel: t('sidebar.worker.type'),
      description: t('sidebar.worker.description'),
      color: 'from-green-500 to-green-600',
    },
  ];

  const handleDragStart = (template: NodeTemplate, event: React.DragEvent) => {
    setDraggingType(template.type);
    event.dataTransfer.setData('application/reactflow/node-type', template.type);
    event.dataTransfer.effectAllowed = 'move';
  };

  const handleDragEnd = () => {
    setDraggingType(null);
  };

  return (
    <>
      {/* Overlay для закрытия */}
      {isOpen && (
        <div
          className="fixed inset-0 bg-black/30 z-40 lg:hidden"
          onClick={onClose}
        />
      )}

      {/* Панель */}
      <div
        className={`fixed right-0 top-0 h-full w-72 bg-[var(--surface)] border-l border-[var(--border)] shadow-xl z-50 transform transition-transform duration-300 ease-in-out ${
          isOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-[var(--border)]">
          <h2 className="text-lg font-bold text-[var(--text)]">{t('sidebar.title')}</h2>
          <button
            onClick={onClose}
            className="text-[var(--text-muted)] hover:text-[var(--text)] transition-colors p-1 rounded"
          >
            ✕
          </button>
        </div>

        {/* Templates */}
        <div className="p-4 space-y-3">
          <p className="text-xs text-[var(--text-muted)] mb-4">
            {t('sidebar.subtitle')}
          </p>

          {templates.map((template) => (
            <div
              key={template.type}
              draggable
              onDragStart={(e) => handleDragStart(template, e)}
              onDragEnd={handleDragEnd}
              className={`cursor-grab active:cursor-grabbing rounded-lg border-2 border-[var(--border)] bg-[var(--background)] hover:border-orange-500 transition-all duration-200 overflow-hidden ${
                draggingType === template.type ? 'opacity-50 scale-95' : 'opacity-100'
              }`}
            >
              {/* Изображение */}
              <div className={`bg-gradient-to-r ${template.color} p-4 flex items-center justify-center`}>
                <img
                  src={template.image}
                  alt={template.label}
                  className="w-24 h-24 object-contain drop-shadow-lg"
                />
              </div>

              {/* Информация */}
              <div className="p-3">
                <div className="flex items-center gap-2 mb-1">
                  <span className="font-bold text-sm text-[var(--text)]">{template.label}</span>
                  <span className="text-xs text-[var(--text-muted)] bg-[var(--surface)] px-2 py-0.5 rounded">
                    {template.typeLabel}
                  </span>
                </div>
                <p className="text-xs text-[var(--text-muted)]">{template.description}</p>
              </div>
            </div>
          ))}

          {/* Подсказка */}
          <div className="mt-6 p-3 bg-[var(--background)] rounded-lg border border-[var(--border)]">
            <div className="text-xs text-[var(--text-muted)] space-y-2">
              <p>💡 <strong>{t('sidebar.tip.title')}</strong></p>
              <ul className="list-disc list-inside space-y-1">
                <li>{t('sidebar.tip.items.0')}</li>
                <li>{t('sidebar.tip.items.1')}</li>
                <li>{t('sidebar.tip.items.2')}</li>
                <li>{t('sidebar.tip.items.3')}</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

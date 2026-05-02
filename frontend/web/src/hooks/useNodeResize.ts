import { useState, useCallback, useEffect } from 'react';
import { useTaskStore } from '../stores/taskStore';

export function useNodeResize(nodeId: string, initialScale: number = 1) {
  const updateNode = useTaskStore((state) => state.updateNode);
  const [isResizing, setIsResizing] = useState(false);
  const [scale, setScale] = useState(initialScale);

  useEffect(() => {
    setScale(initialScale);
  }, [initialScale]);

  const handleResize = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsResizing(true);
  }, []);

  useEffect(() => {
    if (!isResizing) return;

    const handleMouseMove = (e: MouseEvent) => {
      const delta = e.movementX > 0 ? 0.05 : -0.05;
      const newScale = Math.max(0.5, Math.min(2, scale + delta));
      setScale(newScale);
      updateNode(nodeId, { scale: newScale });
    };

    const handleMouseUp = () => {
      setIsResizing(false);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isResizing, scale, nodeId, updateNode]);

  return { scale, handleResize, isResizing };
}
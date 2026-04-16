import { Handle, Position } from '@xyflow/react';
import { Brain } from 'lucide-react';

export function BossNode({ data }: { data: any }) {
  return (
    <div className="bg-gray-900 rounded-xl shadow-lg p-4 min-w-[220px] border border-gray-800">
      <Handle type="target" position={Position.Top} className="!bg-purple-500 !w-2.5 !h-2.5" />
      
      <div className="flex items-center gap-3 mb-3">
        <div className="bg-purple-500/10 p-2 rounded-lg border border-purple-500/20">
          <Brain className="w-5 h-5 text-purple-400" />
        </div>
        <div className="flex-1">
          <h3 className="text-white font-semibold text-sm">{data.name || 'Boss Agent'}</h3>
          <p className="text-gray-400 text-xs">{data.model || 'GPT-4'}</p>
        </div>
      </div>
      
      <div className="flex items-center justify-between bg-gray-800/50 rounded-lg px-3 py-2">
        <span className="text-gray-400 text-xs">Status</span>
        <div className="flex items-center gap-1.5">
          <div className="w-1.5 h-1.5 bg-green-400 rounded-full animate-pulse"></div>
          <span className="text-gray-300 text-xs">Active</span>
        </div>
      </div>
      
      <Handle type="source" position={Position.Bottom} className="!bg-purple-500 !w-2.5 !h-2.5" />
    </div>
  );
}
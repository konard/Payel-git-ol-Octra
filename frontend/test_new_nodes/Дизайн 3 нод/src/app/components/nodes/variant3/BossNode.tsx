import { Handle, Position } from '@xyflow/react';
import { Brain } from 'lucide-react';

export function BossNode({ data }: { data: any }) {
  return (
    <div className="relative bg-slate-800/90 backdrop-blur-sm border border-slate-700 rounded-2xl shadow-lg p-4 min-w-[200px]">
      <Handle type="target" position={Position.Top} className="!bg-purple-500 !w-2.5 !h-2.5" />
      
      <div className="flex items-center gap-3 mb-3">
        <div className="bg-slate-700/50 p-2 rounded-xl border border-slate-600/30">
          <Brain className="w-5 h-5 text-purple-400" />
        </div>
        <div className="flex-1">
          <h3 className="text-white font-semibold text-sm">{data.name || 'Boss Agent'}</h3>
          <p className="text-slate-400 text-xs">{data.model || 'GPT-4'}</p>
        </div>
      </div>
      
      <div className="flex items-center gap-2 bg-slate-700/30 backdrop-blur-sm rounded-lg px-3 py-1.5">
        <div className="w-1.5 h-1.5 bg-green-400 rounded-full animate-pulse"></div>
        <span className="text-slate-300 text-xs">Active</span>
      </div>
      
      <Handle type="source" position={Position.Bottom} className="!bg-purple-500 !w-2.5 !h-2.5" />
    </div>
  );
}
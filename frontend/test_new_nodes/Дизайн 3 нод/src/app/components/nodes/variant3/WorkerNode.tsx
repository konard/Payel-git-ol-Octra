import { Handle, Position } from '@xyflow/react';
import { Cpu } from 'lucide-react';

export function WorkerNode({ data }: { data: any }) {
  return (
    <div className="relative bg-slate-800/90 backdrop-blur-sm border border-slate-700 rounded-2xl shadow-lg p-3.5 min-w-[160px]">
      <Handle type="target" position={Position.Top} className="!bg-emerald-500 !w-2.5 !h-2.5" />
      
      <div className="flex items-center gap-2.5 mb-2.5">
        <div className="bg-slate-700/50 p-1.5 rounded-xl border border-slate-600/30">
          <Cpu className="w-4 h-4 text-emerald-400" />
        </div>
        <div className="flex-1">
          <h3 className="text-white font-semibold text-sm">{data.name || 'Worker Agent'}</h3>
          <p className="text-slate-400 text-xs">{data.model || 'Claude'}</p>
        </div>
      </div>
      
      <div className="flex items-center gap-2 bg-slate-700/30 backdrop-blur-sm rounded-lg px-2.5 py-1.5">
        <div className="w-1.5 h-1.5 bg-green-400 rounded-full animate-pulse"></div>
        <span className="text-slate-300 text-xs">Idle</span>
      </div>
      
      <Handle type="source" position={Position.Bottom} className="!bg-emerald-500 !w-2.5 !h-2.5" />
    </div>
  );
}
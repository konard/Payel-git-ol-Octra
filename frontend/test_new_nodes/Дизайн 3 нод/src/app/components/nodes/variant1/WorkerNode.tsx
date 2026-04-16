import { Handle, Position } from '@xyflow/react';
import { Cpu } from 'lucide-react';

export function WorkerNode({ data }: { data: any }) {
  return (
    <div className="bg-gray-900 rounded-xl shadow-lg p-3.5 min-w-[180px] border border-gray-800">
      <Handle type="target" position={Position.Top} className="!bg-emerald-500 !w-2.5 !h-2.5" />
      
      <div className="flex items-center gap-2.5 mb-3">
        <div className="bg-emerald-500/10 p-1.5 rounded-lg border border-emerald-500/20">
          <Cpu className="w-4 h-4 text-emerald-400" />
        </div>
        <div className="flex-1">
          <h3 className="text-white font-semibold text-sm">{data.name || 'Worker Agent'}</h3>
          <p className="text-gray-400 text-xs">{data.model || 'Claude'}</p>
        </div>
      </div>
      
      <div className="flex items-center justify-between bg-gray-800/50 rounded-lg px-2.5 py-1.5">
        <span className="text-gray-400 text-xs">Status</span>
        <div className="flex items-center gap-1.5">
          <div className="w-1.5 h-1.5 bg-green-400 rounded-full animate-pulse"></div>
          <span className="text-gray-300 text-xs">Idle</span>
        </div>
      </div>
      
      <Handle type="source" position={Position.Bottom} className="!bg-emerald-500 !w-2.5 !h-2.5" />
    </div>
  );
}
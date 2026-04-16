import { Handle, Position } from '@xyflow/react';
import { Cpu } from 'lucide-react';

export function WorkerNode({ data }: { data: any }) {
  return (
    <div className="bg-white/80 backdrop-blur-sm border border-gray-200 rounded-lg shadow-sm p-3.5 min-w-[160px]">
      <Handle type="target" position={Position.Top} className="!bg-emerald-500 !w-2.5 !h-2.5" />
      
      <div className="flex items-center gap-2.5 mb-2.5">
        <div className="bg-gray-100 p-1.5 rounded-lg">
          <Cpu className="w-4 h-4 text-gray-700" />
        </div>
        <div className="flex-1">
          <h3 className="text-gray-900 font-semibold text-sm">{data.name || 'Worker Agent'}</h3>
          <p className="text-gray-500 text-xs">{data.model || 'Claude'}</p>
        </div>
      </div>
      
      <div className="flex items-center gap-2">
        <div className="w-2 h-2 bg-green-500 rounded-full"></div>
        <span className="text-gray-600 text-xs">Idle</span>
      </div>
      
      <Handle type="source" position={Position.Bottom} className="!bg-emerald-500 !w-2.5 !h-2.5" />
    </div>
  );
}
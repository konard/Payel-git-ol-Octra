import { Handle, Position } from '@xyflow/react';
import { Brain } from 'lucide-react';

export function BossNode({ data }: { data: any }) {
  return (
    <div className="bg-white/80 backdrop-blur-sm border border-gray-200 rounded-lg shadow-md p-4 min-w-[200px]">
      <Handle type="target" position={Position.Top} className="!bg-purple-500 !w-2.5 !h-2.5" />
      
      <div className="flex items-center gap-3 mb-3">
        <div className="bg-gray-100 p-2 rounded-lg">
          <Brain className="w-5 h-5 text-gray-700" />
        </div>
        <div className="flex-1">
          <h3 className="text-gray-900 font-semibold text-sm">{data.name || 'Boss Agent'}</h3>
          <p className="text-gray-500 text-xs">{data.model || 'GPT-4'}</p>
        </div>
      </div>
      
      <div className="flex items-center gap-2">
        <div className="w-2 h-2 bg-green-500 rounded-full"></div>
        <span className="text-gray-600 text-xs">Active</span>
      </div>
      
      <Handle type="source" position={Position.Bottom} className="!bg-purple-500 !w-2.5 !h-2.5" />
    </div>
  );
}
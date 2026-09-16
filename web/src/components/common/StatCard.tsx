import React from 'react';
import { LucideIcon } from 'lucide-react';

interface StatCardProps {
  label: string;
  value: string | number;
  icon: LucideIcon;
  color?: string;
  change?: string;
  isPositive?: boolean;
}

export const StatCard: React.FC<StatCardProps> = ({
  label,
  value,
  icon: Icon,
  color = 'text-cyan-400',
  change,
  isPositive = true,
}) => {
  return (
    <div className="liquid-glass-card p-4 sm:p-5 rounded-2xl sm:rounded-3xl relative overflow-hidden group">
      <div className="flex items-center justify-between gap-2 mb-3">
        <span className="text-xs font-semibold text-slate-400 truncate">{label}</span>
        <div className={`w-8 h-8 rounded-xl bg-white/[0.04] border border-white/10 flex items-center justify-center ${color} group-hover:scale-110 transition-transform`}>
          <Icon size={16} />
        </div>
      </div>

      <div className="flex items-baseline justify-between gap-2">
        <div className="text-xl sm:text-2xl font-black text-white tracking-tight">{value}</div>
        {change && (
          <span className={`text-[11px] font-bold px-1.5 py-0.5 rounded-lg border ${
            isPositive 
              ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400' 
              : 'bg-red-500/10 border-red-500/20 text-red-400'
          }`}>
            {change}
          </span>
        )}
      </div>
    </div>
  );
};

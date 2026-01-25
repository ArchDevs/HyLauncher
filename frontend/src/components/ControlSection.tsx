import React from "react";
import { motion } from "framer-motion";
import { FolderOpen, Activity, Settings, Trash } from "lucide-react";

interface ControlSectionProps {
  onPlay: () => void;
  isDownloading: boolean;
  progress: number;
  status: string;
  speed: string; // Added
  downloaded: number; // Added
  total: number; // Added
  currentFile: string; // Added
  actions: {
    openFolder: () => void;
    showDiagnostics: () => void;
    showDelete: () => void;
  };
}

export const ControlSection: React.FC<ControlSectionProps> = ({
  onPlay,
  isDownloading,
  progress,
  status,
  speed,
  downloaded,
  total,
  currentFile,
  actions,
}) => {
  // Your original formatting helper
  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  return (
    <div className="w-full flex items-end gap-8">
      <div className="w-[294px] flex flex-col gap-3">
        <div className="flex gap-[10px]">
          <NavBtn
            onClick={actions.openFolder}
            icon={<FolderOpen size={20} />}
          />
          <NavBtn
            onClick={actions.showDiagnostics}
            icon={<Activity size={20} />}
          />
          <NavBtn icon={<Settings size={20} />} />
          <NavBtn onClick={actions.showDelete} icon={<Trash size={20} />} />
        </div>
        <motion.button
          whileTap={isDownloading ? {} : { scale: 0.98 }}
          onClick={onPlay}
          disabled={isDownloading}
          className={`w-full h-[94px] font-unbounded bg-[#090909]/[0.55] backdrop-blur-xl text-white font-black text-4xl tracking-tighter rounded-[14px] border border-[#FFA845]/[0.10] shadow-lg disabled:opacity-50 ${
            isDownloading ? "cursor-not-allowed" : "cursor-pointer"
          }`}
        >
          {isDownloading ? "INSTALL..." : "PLAY"}
        </motion.button>
      </div>

      <div className="flex-1 flex flex-col gap-4 pb-1">
        <div className="flex justify-between items-end">
          <div className="flex items-baseline gap-4">
            <span className="text-[36px] font-unbounded font-[500] tracking-tighter">
              {Math.round(progress)}%
            </span>
            <span className="text-[11px] text-gray-400 uppercase font-bold tracking-widest opacity-70">
              {status}
            </span>
          </div>

          {/* Re-added speed and total size labels */}
          <div className="text-[11px] text-gray-400 font-mono">
            {speed && total > 0
              ? `${speed} • ${formatBytes(downloaded)} / ${formatBytes(total)}`
              : currentFile || "Ready"}
          </div>
        </div>
        <div className="h-2 w-full bg-white/5 rounded-full overflow-hidden border border-white/5">
          <motion.div
            animate={{ width: `${progress}%` }}
            className="h-full bg-white progress-glow"
          />
        </div>
      </div>
    </div>
  );
};

const NavBtn = ({ icon, onClick }: { icon: any; onClick?: () => void }) => (
  <button
    onClick={onClick}
    className="w-[66px] h-[42px] cursor-pointer flex items-center justify-center bg-[#090909]/[0.55] backdrop-blur-xl border border-[#FFA845]/[0.10] rounded-[14px] hover:bg-[#FFA845]/[0.05] hover:border-[#FFA845]/[0.30] transition-all text-gray-400 hover:text-white"
  >
    {icon}
  </button>
);

import React from "react";
import { motion } from "framer-motion";
import {
  FolderOpen,
  Activity,
  Settings,
  Trash,
  ArrowUpCircle,
  RefreshCcw,
} from "lucide-react";
import BackgroundImage from "./BackgroundImage";

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
  updateAvailable: boolean;
  onUpdate: () => void;
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
  updateAvailable,
  onUpdate,
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
    <div className="w-full flex items-end gap-[20px] ml-[48px]">
      <div className="w-[280px] flex flex-col gap-[12px]">
        <div className="w-[280px] h-[120px]  bg-[#090909]/[0.55] backdrop-blur-[12] border border-[#FFA845]/[0.10] rounded-[14px]"></div>
        {updateAvailable && (
          <button
            onClick={onUpdate}
            className="cursor-pointer hover:scale-102 w-[280px] h-[40px] bg-[#090909]/[0.55] backdrop-blur-[12] border border-[#FFA845]/[0.10] rounded-[12px] px-[12px] flex items-center justify-between"
          >
            <span className="text-[16px] text-[#CCD9E0]/[0.90] font-[Mazzard] tracking-[-3%]">
              Доступно обновление
            </span>
            <span className="text-[#CCD9E0]/[0.90] transition-transform flex items-center justify-center">
              <RefreshCcw size={16} />
            </span>
          </button>
        )}

        {/* 
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
        </div> */}
        <motion.button
          whileTap={isDownloading ? {} : { scale: 0.98 }}
          onClick={onPlay}
          disabled={isDownloading}
          className={`w-[280px] h-[100px] font-[Unbounded] font-[600] text-[32px] text-[#CCD9E0]/[0.90] bg-[#090909]/[0.55] backdrop-blur-[12] rounded-[14px] border border-[#FFA845]/[0.10] shadow-lg disabled:opacity-50 hover:scale-102 ${
            isDownloading ? "cursor-not-allowed" : "cursor-pointer"
          }`}
        >
          {isDownloading ? "INSTALL..." : "PLAY"}
        </motion.button>
      </div>

      <div className="flex-1 flex flex-col gap-[6px] pb-1">
        <div className="flex justify-between items-end">
          <div className="tracking-[-3%]  flex items-baseline gap-[20px] ">
            <span className="text-[34px] text-[#CCD9E0]/[0.90] font-[Unbounded] font-[500]">
              {Math.round(progress)}%
            </span>
            <span className="text-[16px] text-[#CCD9E0]/[0.30] font-[Mazzard]">
              {status}
            </span>
          </div>

          {/* Re-added speed and total size labels */}
          <div className="text-[14px] text-[#CCD9E0]/[0.30] font-[MazzardM-Medium] max-w-[200px] truncate mr-[48px]">
            {speed && total > 0
              ? `${speed} • ${formatBytes(downloaded)} / ${formatBytes(total)}`
              : currentFile || "Ready"}
          </div>
        </div>
        <div className="h-[7px] w-[852px] bg-[#090909]/[0.10] rounded-full overflow-hidden border border-[#FFA845]/[0.10] border-[0.5px]">
          <motion.div
            animate={{ width: `${progress}%` }}
            className="h-full bg-[#CCD9E0]/[0.90] progress-glow"
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

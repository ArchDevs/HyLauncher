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
import { useTranslation } from "../i18n";

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
  latestNews?: {
    title: string;
    dest_url: string;
    description: string;
    image_url: string;
  };
  onOpenNews?: (url: string) => void;
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
  latestNews,
  onOpenNews,
}) => {
  const { t } = useTranslation();

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
        {latestNews ? (
          <button
            onClick={() => onOpenNews?.(latestNews.dest_url)}
            className="w-[280px] h-[120px] bg-[#090909]/[0.55] backdrop-blur-[12px] border border-[#FFA845]/[0.10] rounded-[14px] overflow-hidden group block cursor-pointer text-left"
          >
            <div className="relative w-full h-full">
              <img
                key={latestNews.title}
                src={
                  latestNews.image_url.startsWith("http")
                    ? latestNews.image_url
                    : `https://launcher.hytale.com/launcher-feed/release/${latestNews.image_url}`
                }
                alt={latestNews.title}
                className="absolute inset-0 w-full h-full object-cover opacity-40 group-hover:opacity-60 transition-opacity duration-500"
              />
              <div className="absolute inset-0 bg-gradient-to-t from-black/80 to-transparent" />
              <div className="absolute bottom-0 left-0 p-3 w-full">
                <h3 className="text-[14px] font-[Unbounded] font-[500] text-[#CCD9E0]/[0.90] line-clamp-1 mb-0.5">
                  {latestNews.title}
                </h3>
                <p className="text-[11px] font-[Mazzard] text-[#CCD9E0]/[0.40] line-clamp-2 leading-tight">
                  {latestNews.description}
                </p>
              </div>
            </div>
          </button>
        ) : (
          <div className="w-[280px] h-[120px] bg-[#090909]/[0.55] backdrop-blur-[12px] border border-[#FFA845]/[0.10] rounded-[14px] animate-pulse flex items-center justify-center">
            <span className="text-[12px] text-[#CCD9E0]/[0.20] font-[Mazzard]">Loading news...</span>
          </div>
        )}
        {updateAvailable && (
          <button
            onClick={onUpdate}
            className="cursor-pointer hover:scale-102 w-[280px] h-[40px] bg-[#090909]/[0.55] backdrop-blur-[12px] border border-[#FFA845]/[0.10] rounded-[12px] px-[12px] flex items-center justify-between"
          >
            <span className="text-[16px] text-[#CCD9E0]/[0.90] font-[Mazzard] tracking-[-3%]">
              {t.control.updateAvailable}
            </span>
            <span className="text-[#CCD9E0]/[0.90] transition-transform flex items-center justify-center">
              <RefreshCcw size={16} />
            </span>
          </button>
        )}
        <motion.button
          style={{
            willChange: "transform",
            transform: "translateZ(0)",
            backfaceVisibility: "hidden",
          }}
          whileTap={isDownloading ? {} : { scale: 0.98 }}
          onClick={onPlay}
          disabled={isDownloading}
          transition={{
            type: "spring",
            stiffness: 400,
            damping: 25,
          }}
          className={`w-[280px] h-[100px] font-[Unbounded] font-[600] text-[32px] text-[#CCD9E0]/[0.90] bg-[#090909]/[0.55] backdrop-blur-[12px] rounded-[14px] border border-[#FFA845]/[0.10] shadow-lg disabled:opacity-50 hover:scale-102 ${
            isDownloading ? "cursor-not-allowed" : "cursor-pointer"
          }`}
        >
          {isDownloading ? t.common.install : t.common.play}
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
          <div className="text-[14px] text-[#CCD9E0]/[0.30] font-[MazzardM-Medium] text-right break-words min-w-0 flex-1 mr-[48px]">
            {speed && total > 0
              ? `${speed} • ${formatBytes(downloaded)} / ${formatBytes(total)}`
              : currentFile || t.common.ready}
          </div>
        </div>
        <div className="h-[7px] w-[852px] bg-[#090909]/[0.10] rounded-full overflow-hidden border border-[#FFA845]/[0.10] border-[0.2px]">
          <motion.div
            style={{
              willChange: "width",
              transform: "translateZ(0)",
              backfaceVisibility: "hidden",
            }}
            animate={{ width: `${progress}%` }}
            transition={{
              type: "tween",
              ease: "linear",
              duration: 0.1,
            }}
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

import React, { useState } from "react";
import { motion } from "framer-motion";
import { X, Copy, Play, Check } from "lucide-react";
import { service } from "../../wailsjs/go/models";
import { useTranslation } from "../i18n";

// Use the generated type
type ServerWithFullUrls = service.ServerWithUrls;

interface ServerModalProps {
  server: ServerWithFullUrls | null;
  isOpen: boolean;
  onClose: () => void;
  onPlay?: (serverIP: string) => void;
}

export const ServerModal: React.FC<ServerModalProps> = ({
  server,
  isOpen,
  onClose,
  onPlay,
}) => {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);

  if (!isOpen || !server) return null;

  const handleCopyIp = async () => {
    try {
      await navigator.clipboard.writeText(server.ip);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy IP:", err);
    }
  };

  const handlePlay = () => {
    if (onPlay) {
      onPlay(server.ip);
    }
    onClose();
  };

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <motion.div
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.9, opacity: 0 }}
        transition={{ type: "spring", damping: 20, stiffness: 300 }}
        onClick={(e) => e.stopPropagation()}
        className="relative w-[480px] bg-[#090909]/75 backdrop-blur-[12px] rounded-[20px] border border-[#7C7C7C]/10 overflow-hidden"
      >
        {/* Close Button */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 z-10 p-2 text-white/50 hover:text-white transition-colors cursor-pointer bg-[#090909]/50 rounded-[10px] border border-[#7C7C7C]/10"
        >
          <X size={18} strokeWidth={1.6} />
        </button>

        {/* Banner Image */}
        <div className="relative w-full h-[180px] overflow-hidden">
          {server.banner_url ? (
            <img
              src={server.banner_url}
              alt={server.name}
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="w-full h-full bg-[#090909]/55 flex items-center justify-center">
              <span className="text-white/30 font-[Unbounded]">No Banner</span>
            </div>
          )}
          {/* Gradient overlay for smooth transition to content */}
          <div className="absolute bottom-0 left-0 right-0 h-[60px] bg-gradient-to-t from-[#090909]/75 to-transparent" />
        </div>

        {/* Content */}
        <div className="px-6 pb-6 -mt-6 relative">
          {/* Logo and Name Row */}
          <div className="flex items-start gap-4 mb-4">
            {/* Logo */}
            {server.logo_url && (
              <div className="w-[80px] h-[80px] rounded-[14px] overflow-hidden border border-[#7C7C7C]/10 bg-[#090909]/55 flex-shrink-0">
                <img
                  src={server.logo_url}
                  alt={`${server.name} logo`}
                  className="w-full h-full object-cover"
                />
              </div>
            )}

            {/* Name and IP */}
            <div className="flex-1 pt-2">
              <h2 className="text-[20px] font-[Unbounded] font-semibold text-white/90 tracking-[-0.02em]">
                {server.name}
              </h2>
              <p className="text-[14px] font-[Mazzard] text-[#FFA845]/80 mt-1">
                {server.ip}
              </p>
            </div>
          </div>

          {/* Description */}
          <p className="text-[15px] font-[Mazzard] text-white/60 leading-[1.5] mb-6">
            {server.description}
          </p>

          {/* Action Buttons */}
          <div className="flex gap-3">
            {/* Copy IP Button */}
            <button
              onClick={handleCopyIp}
              className="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-[#090909]/55 border border-[#7C7C7C]/10 rounded-[14px] hover:bg-[#090909]/70 hover:border-[#FFA845]/20 transition-all cursor-pointer group"
            >
              {copied ? (
                <>
                  <Check size={18} className="text-green-400" strokeWidth={1.6} />
                  <span className="text-[15px] font-[Mazzard] font-semibold text-green-400 tracking-[-0.02em]">
                    {t.modals?.server?.copied || "Copied!"}
                  </span>
                </>
              ) : (
                <>
                  <Copy size={18} className="text-[#CCD9E0]/80 group-hover:text-[#FFA845] transition-colors" strokeWidth={1.6} />
                  <span className="text-[15px] font-[Mazzard] font-semibold text-[#CCD9E0]/80 group-hover:text-white tracking-[-0.02em]">
                    {t.modals?.server?.copyIp || "Copy IP"}
                  </span>
                </>
              )}
            </button>

            {/* Play Button */}
            <button
              onClick={handlePlay}
              className="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-[#FFA845]/20 border border-[#FFA845]/30 rounded-[14px] hover:bg-[#FFA845]/30 hover:border-[#FFA845]/50 transition-all cursor-pointer group"
            >
              <Play size={18} className="text-[#FFA845] fill-[#FFA845]" strokeWidth={1.6} />
              <span className="text-[15px] font-[Mazzard] font-semibold text-[#FFA845] tracking-[-0.02em]">
                {t.modals?.server?.play || "Play"}
              </span>
            </button>
          </div>
        </div>
      </motion.div>
    </motion.div>
  );
};

export default ServerModal;

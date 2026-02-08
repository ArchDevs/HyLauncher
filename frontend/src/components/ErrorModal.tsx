import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { AlertCircle, Copy, ChevronDown, ChevronUp, X, Headset } from 'lucide-react';
import { useTranslation } from '../i18n';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';

interface AppError {
  type: string;
  message: string;
  technical: string;
  timestamp?: string;
  stack?: string;
}

interface ErrorModalProps {
  error: AppError;
  onClose: () => void;
}

export const ErrorModal: React.FC<ErrorModalProps> = ({ error, onClose }) => {
  const { t } = useTranslation();
  const [showTechnical, setShowTechnical] = useState(false);
  const [copied, setCopied] = useState(false);

  const copyErrorDetails = () => {
    const details = `
Error Type: ${error.type}
Time: ${error.timestamp || new Date().toISOString()}
Message: ${error.message}
Technical: ${error.technical}
${error.stack ? `Stack:\n${error.stack}` : ''}
    `.trim();

    navigator.clipboard.writeText(details);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
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
        onClick={(e) => e.stopPropagation()}
        className="relative w-[550px] bg-[#090909]/75 backdrop-blur-[6px] rounded-[14px] border border-[#FFA845]/10 overflow-hidden"
      >
        {/* Header */}
        <div className="relative px-5 pt-5 pb-0 flex items-start gap-4">
          {/* Warning Icon Box */}
          <div className="flex-shrink-0 w-12 h-12 flex items-center justify-center bg-[#090909]/55 border border-[#7C7C7C]/10 rounded-[14px]">
            <div className="relative">
              <AlertCircle size={18} className="text-[#FFA845]/90" strokeWidth={1.6} />
              <div className="absolute top-1/2 left-1/2 w-[1px] h-[3px] bg-[#FFA845]/90 -translate-x-1/2" />
              <div className="absolute bottom-[2px] left-1/2 w-[1px] h-[1px] bg-[#FFA845]/90 -translate-x-1/2" />
            </div>
          </div>

          {/* Title & Error Type */}
          <div className="flex-1 pt-1">
            <h3 className="text-[16px] font-[Unbounded] font-semibold text-white/90 tracking-[-0.03em]">
              {t.modals?.error?.title || "An error occurred"}
            </h3>
            <p className="text-[14px] font-[Mazzard] font-medium text-white/25 mt-1 tracking-[-0.03em]">
              {error.type}
            </p>
          </div>

          {/* Close Button */}
          <button
            onClick={onClose}
            className="p-1 text-white/50 hover:text-white transition-colors mt-1 cursor-pointer"
          >
            <X size={18} strokeWidth={1.6} />
          </button>
        </div>

        {/* Divider */}
        <div className="relative h-[1px] bg-[#FFA845]/10 mt-5 mx-0" />

        {/* Content */}
        <div className="relative px-5 py-6 space-y-4">
          {/* Error Message */}
          <p className="text-[16px] font-[Mazzard] font-medium text-white/50 tracking-[-0.03em]">
            {error.message}
          </p>

          {/* Suggestion Box */}
          <button
            onClick={() => BrowserOpenURL("https://t.me/hylauncher_bot")}
            className="cursor-pointer w-full flex items-center justify-between gap-4 px-4 py-4 bg-[#090909]/55 border border-[#7C7C7C]/10 rounded-[14px] hover:bg-[#090909]/70 transition-colors text-left"
          >
            <span className="text-[16px] font-[Mazzard] font-semibold text-[#CCD9E0]/90 tracking-[-0.03em]">
              {t.modals?.error?.suggestion || "Please report this issue if it persists."}
            </span>
            <Headset size={18} className="text-[#CCD9E0]/90 flex-shrink-0" strokeWidth={1.6} />
          </button>

          {/* Technical Details Toggle */}
          {error.technical && (
            <>
              <button
                onClick={() => setShowTechnical(!showTechnical)}
                className="cursor-pointer w-full flex items-center justify-between px-4 py-3 bg-[#090909]/55 border border-[#7C7C7C]/10 rounded-[14px] hover:bg-[#090909]/70 transition-colors"
              >
                <span className="text-[16px] font-[Mazzard] font-semibold text-[#CCD9E0]/90 tracking-[-0.03em]">
                  {t.modals?.error?.technicalDetails || "Technical details"}
                </span>
                {showTechnical ? (
                  <ChevronUp size={16} className="text-[#CCD9E0]/90" strokeWidth={1.6} />
                ) : (
                  <ChevronDown size={16} className="text-[#CCD9E0]/90" strokeWidth={1.6} />
                )}
              </button>

              <AnimatePresence>
                {showTechnical && (
                  <motion.div
                    initial={{ height: 0, opacity: 0 }}
                    animate={{ height: 'auto', opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    className="overflow-hidden"
                  >
                    <div className="px-4 py-3 bg-[#050505]/75 border border-[#7C7C7C]/10 rounded-[14px]">
                      <pre className="text-[16px] font-[Mazzard] font-semibold text-[#CCD9E0]/50 tracking-[-0.03em] whitespace-pre-wrap break-all">
                        {error.technical}
                      </pre>
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>
            </>
          )}
        </div>

        {/* Footer - Copy Button */}
        <div className="relative px-5 pb-5">
          <button
            onClick={copyErrorDetails}
            className="cursor-pointer w-full flex items-center justify-center gap-2 px-4 py-3 bg-[#090909]/55 border border-[#FFA845]/10 rounded-[14px] hover:bg-[#090909]/70 transition-colors"
          >
            <span className="text-[16px] font-[Mazzard] font-semibold text-[#CCD9E0]/90 tracking-[-0.03em]">
              {copied 
                ? (t.modals?.error?.copied || "Copied!") 
                : (t.modals?.error?.copyError || "Copy error")
              }
            </span>
            <Copy size={16} className="text-[#CCD9E0]/90" strokeWidth={1.6} />
          </button>
        </div>
      </motion.div>
    </motion.div>
  );
};
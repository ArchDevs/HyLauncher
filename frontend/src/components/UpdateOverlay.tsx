import React from 'react';
import { motion } from 'framer-motion';
import { DownloadCloud } from 'lucide-react';
import { useTranslation } from '../i18n';

interface UpdateOverlayProps {
  progress: number;
  downloaded: number;
  total: number;
}

export const UpdateOverlay: React.FC<UpdateOverlayProps> = ({ progress, downloaded, total }) => {
  const { t } = useTranslation();
  const toMB = (bytes: number) => (bytes / 1024 / 1024).toFixed(1);

  return (
    <motion.div 
      initial={{ opacity: 0 }} 
      animate={{ opacity: 1 }}
      className="absolute inset-0 z-[100] bg-[#090909]/95 backdrop-blur-2xl flex flex-col items-center justify-center p-20 text-center"
    >
      <DownloadCloud size={80} className="text-[#FFA845] mb-8 animate-bounce" />
      <h1 className="text-5xl font-black mb-4 tracking-tighter italic text-white">{t.modals.update.title}</h1>
      <p className="text-gray-400 mb-12 max-w-md text-lg font-medium">
        {t.modals.update.message}
      </p>
      
      <div className="w-full max-w-2xl">
        <div className="flex justify-between mb-4 items-end">
            <span className="text-6xl font-black italic tracking-tighter text-white">
                {Math.round(progress)}%
            </span>
            <span className="text-sm font-mono text-gray-500 uppercase tracking-widest bg-white/5 px-3 py-1 rounded-full">
                {toMB(downloaded)}MB / {toMB(total)}MB
            </span>
        </div>
        
        <div className="h-3 w-full bg-white/5 rounded-full overflow-hidden border border-white/10 p-[2px]">
          <motion.div 
            initial={{ width: 0 }}
            animate={{ width: `${progress}%` }} 
            className="h-full bg-white progress-glow"
          />
        </div>
      </div>
    </motion.div>
  );
};
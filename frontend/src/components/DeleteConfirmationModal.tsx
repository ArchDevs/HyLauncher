import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';

interface DeleteConfirmationModalProps {
  onConfirm: () => void;
  onCancel: () => void;
}

export const DeleteConfirmationModal: React.FC<DeleteConfirmationModalProps> = ({
  onConfirm,
  onCancel,
}) => {
  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 bg-black/70 backdrop-blur-sm z-50 flex items-center justify-center p-4"
      >
        <motion.div
          initial={{ scale: 0.85, y: 20, opacity: 0 }}
          animate={{ scale: 1, y: 0, opacity: 1 }}
          exit={{ scale: 0.85, y: 20, opacity: 0 }}
          transition={{ type: "spring", damping: 20, stiffness: 300 }}
          className="bg-[#0f0f0f] border border-[#FFA845]/20 rounded-2xl p-8 max-w-md w-full shadow-2xl"
        >
          <h2 className="text-2xl font-bold text-white mb-4">Вы уверены?</h2>

          <p className="text-gray-300 mb-8 leading-relaxed">
            Вы точно хотите удалить игру?<br />
            <span className="text-red-400 font-medium">
              Это действие удалит все файлы игры без возможности восстановления!
            </span>
          </p>

          <div className="flex gap-4 justify-end">
            <button
              onClick={onCancel}
              className="px-6 py-3 bg-[#1a1a1a] hover:bg-[#222] text-gray-300 rounded-lg transition-colors border border-white/10"
            >
              Отмена
            </button>
            <button
              onClick={onConfirm}
              className="px-6 py-3 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg transition-colors shadow-lg shadow-red-900/30"
            >
              Удалить всё
            </button>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
};
import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { AlertCircle, Copy, ChevronDown, ChevronUp, X } from 'lucide-react';
import { useTranslation } from '../i18n';

interface AppError {
  type: string;
  message: string;
  technical: string;
  timestamp: string;
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
Time: ${error.timestamp}
Message: ${error.message}
Technical: ${error.technical}
${error.stack ? `Stack:\n${error.stack}` : ''}
    `.trim();

    navigator.clipboard.writeText(details);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const getErrorColor = (type: string) => {
    switch (type) {
      case 'NETWORK':
        return '#FFA845';
      case 'FILESYSTEM':
        return '#FF6B6B';
      case 'VALIDATION':
        return '#FFD93D';
      case 'GAME':
        return '#FF8787';
      default:
        return '#FFA845';
    }
  };

  const getSuggestion = (type: string) => {
    switch (type) {
      case 'NETWORK':
        return t.modals.error.suggestions.network;
      case 'FILESYSTEM':
        return t.modals.error.suggestions.filesystem;
      case 'VALIDATION':
        return t.modals.error.suggestions.validation;
      case 'GAME':
        return t.modals.error.suggestions.game;
      default:
        return t.modals.error.suggestions.default;
    }
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
        className="w-full max-w-md bg-[#090909]/95 backdrop-blur-xl rounded-2xl border border-white/10 overflow-hidden shadow-2xl"
        style={{ borderColor: getErrorColor(error.type) + '40' }}
      >
        {/* Header */}
        <div
          className="p-6 border-b border-white/10"
          style={{
            background: `linear-gradient(135deg, ${getErrorColor(error.type)}20 0%, transparent 100%)`,
          }}
        >
          <div className="flex items-start justify-between">
            <div className="flex items-start gap-3">
              <div
                className="p-2 rounded-lg"
                style={{ backgroundColor: getErrorColor(error.type) + '20' }}
              >
                <AlertCircle size={24} style={{ color: getErrorColor(error.type) }} />
              </div>
              <div>
                <h3 className="text-lg font-bold text-white">{t.modals.error.title}</h3>
                <p className="text-xs text-gray-400 mt-1">{error.type}</p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="p-1 hover:bg-white/10 rounded-lg transition-colors"
            >
              <X size={20} className="text-gray-400" />
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="p-6 space-y-4">
          {/* User Message */}
          <div>
            <p className="text-sm text-gray-200 leading-relaxed">{error.message}</p>
          </div>

          {/* Suggestion */}
          <div className="bg-white/5 rounded-lg p-3 border border-white/5">
            <p className="text-xs text-gray-300">{getSuggestion(error.type)}</p>
          </div>

          {/* Technical Details Toggle */}
          {error.technical && (
            <div>
              <button
                onClick={() => setShowTechnical(!showTechnical)}
                className="w-full flex items-center justify-between p-3 bg-white/5 hover:bg-white/10 rounded-lg border border-white/5 transition-colors"
              >
                <span className="text-xs font-medium text-gray-300">{t.modals.error.technicalDetails}</span>
                {showTechnical ? (
                  <ChevronUp size={16} className="text-gray-400" />
                ) : (
                  <ChevronDown size={16} className="text-gray-400" />
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
                    <div className="mt-2 p-3 bg-black/50 rounded-lg border border-white/5">
                      <pre className="text-xs text-gray-400 font-mono whitespace-pre-wrap break-all">
                        {error.technical}
                      </pre>
                      {error.stack && (
                        <details className="mt-2">
                          <summary className="text-xs text-gray-500 cursor-pointer hover:text-gray-400">
                            {t.modals.error.stackTrace}
                          </summary>
                          <pre className="text-xs text-gray-500 font-mono mt-2 whitespace-pre-wrap">
                            {error.stack}
                          </pre>
                        </details>
                      )}
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          )}
        </div>

        {/* Actions */}
        <div className="p-6 pt-0 flex gap-2">
          <button
            onClick={copyErrorDetails}
            className="flex-1 px-4 py-2 bg-white/5 hover:bg-white/10 rounded-lg border border-white/5 transition-colors flex items-center justify-center gap-2"
          >
            <Copy size={14} className="text-gray-400" />
            <span className="text-xs font-medium text-gray-300">
              {copied ? t.common.copied : t.common.copy}
            </span>
          </button>
          <button
            onClick={onClose}
            className="flex-1 px-4 py-2 rounded-lg border transition-colors"
            style={{
              backgroundColor: getErrorColor(error.type) + '20',
              borderColor: getErrorColor(error.type) + '40',
            }}
          >
            <span className="text-xs font-medium text-white">{t.common.close}</span>
          </button>
        </div>
      </motion.div>
    </motion.div>
  );
};
import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Activity, X, Copy, Download, RefreshCw, CheckCircle, XCircle, AlertCircle } from 'lucide-react';

interface DiagnosticsModalProps {
  onClose: () => void;
  onRunDiagnostics: () => Promise<DiagnosticReport>;
  onSaveDiagnostics: () => Promise<string>;
}

interface DiagnosticReport {
  timestamp: string;
  app_version: string;
  platform: {
    os: string;
    arch: string;
    go_version: string;
    num_cpu: number;
  };
  connectivity: {
    can_reach_game_server: boolean;
    game_server_error?: string;
    response_time_ms: number;
  };
  local_installation: {
    game_installed: boolean;
    current_version: number;
    install_path: string;
    jre_installed: boolean;
    butler_installed: boolean;
  };
  server_versions: {
    latest_version: number;
    found_versions: boolean;
    checked_urls?: string[];
    error?: string;
  };
  disk_space: {
    install_directory: string;
    error?: string;
  };
}

export const DiagnosticsModal: React.FC<DiagnosticsModalProps> = ({
  onClose,
  onRunDiagnostics,
  onSaveDiagnostics,
}) => {
  const [report, setReport] = useState<DiagnosticReport | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);
  const [saved, setSaved] = useState(false);

  const runDiagnostics = async () => {
    setLoading(true);
    try {
      const result = await onRunDiagnostics();
      setReport(result);
    } catch (err) {
      console.error('Diagnostics failed:', err);
    } finally {
      setLoading(false);
    }
  };

  const copyReport = () => {
    if (!report) return;
    
    const text = formatReportAsText(report);
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const saveReport = async () => {
    try {
      const path = await onSaveDiagnostics();
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch (err) {
      console.error('Failed to save diagnostics:', err);
    }
  };

  const formatReportAsText = (report: DiagnosticReport): string => {
    return `=== HyLauncher Diagnostic Report ===
Generated: ${new Date(report.timestamp).toLocaleString()}
App Version: ${report.app_version}

--- Platform ---
OS: ${report.platform.os}
Architecture: ${report.platform.arch}
CPUs: ${report.platform.num_cpu}

--- Connectivity ---
Game Server: ${report.connectivity.can_reach_game_server ? '✓ Reachable' : '✗ Unreachable'}
Response Time: ${report.connectivity.response_time_ms}ms
${report.connectivity.game_server_error ? `Error: ${report.connectivity.game_server_error}` : ''}

--- Local Installation ---
Game Installed: ${report.local_installation.game_installed ? 'Yes' : 'No'}
Current Version: ${report.local_installation.current_version}
JRE Installed: ${report.local_installation.jre_installed ? 'Yes' : 'No'}
Butler Installed: ${report.local_installation.butler_installed ? 'Yes' : 'No'}
Install Path: ${report.local_installation.install_path}

--- Server Versions ---
Latest Version: ${report.server_versions.latest_version}
Versions Found: ${report.server_versions.found_versions ? 'Yes' : 'No'}
${report.server_versions.error ? `Error: ${report.server_versions.error}` : ''}

${report.server_versions.checked_urls ? `Sample URLs:\n${report.server_versions.checked_urls.join('\n')}` : ''}
`;
  };

  const StatusIcon: React.FC<{ status: boolean }> = ({ status }) => {
    if (status) {
      return <CheckCircle size={16} className="text-green-400" />;
    }
    return <XCircle size={16} className="text-red-400" />;
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
        className="w-full max-w-2xl bg-[#090909]/95 backdrop-blur-xl rounded-2xl border border-[#FFA845]/20 overflow-hidden shadow-2xl max-h-[90vh] flex flex-col"
      >
        {/* Header */}
        <div className="p-6 border-b border-white/10 bg-gradient-to-r from-[#FFA845]/10 to-transparent">
          <div className="flex items-start justify-between">
            <div className="flex items-start gap-3">
              <div className="p-2 rounded-lg bg-[#FFA845]/20">
                <Activity size={24} className="text-[#FFA845]" />
              </div>
              <div>
                <h3 className="text-lg font-bold text-white">System Diagnostics</h3>
                <p className="text-xs text-gray-400 mt-1">Check system status and connectivity</p>
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
        <div className="flex-1 overflow-y-auto p-6">
          {!report ? (
            <div className="flex flex-col items-center justify-center py-12 gap-4">
              <Activity size={48} className="text-[#FFA845]/50" />
              <p className="text-sm text-gray-300 text-center">
                Run diagnostics to check your system status,<br />
                connectivity, and installation.
              </p>
              <button
                onClick={runDiagnostics}
                disabled={loading}
                className="mt-4 px-6 py-3 bg-[#FFA845]/20 hover:bg-[#FFA845]/30 border border-[#FFA845]/40 rounded-lg transition-colors disabled:opacity-50"
              >
                <span className="text-sm font-medium text-white flex items-center gap-2">
                  {loading ? (
                    <>
                      <RefreshCw size={16} className="animate-spin" />
                      Running...
                    </>
                  ) : (
                    <>
                      <Activity size={16} />
                      Run Diagnostics
                    </>
                  )}
                </span>
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              {/* Platform Info */}
              <div className="bg-white/5 rounded-lg p-4 border border-white/5">
                <h4 className="text-sm font-bold text-white mb-3">Platform</h4>
                <div className="grid grid-cols-2 gap-3 text-xs">
                  <div>
                    <span className="text-gray-400">OS:</span>
                    <span className="text-gray-200 ml-2">{report.platform.os}</span>
                  </div>
                  <div>
                    <span className="text-gray-400">Arch:</span>
                    <span className="text-gray-200 ml-2">{report.platform.arch}</span>
                  </div>
                  <div>
                    <span className="text-gray-400">CPUs:</span>
                    <span className="text-gray-200 ml-2">{report.platform.num_cpu}</span>
                  </div>
                  <div>
                    <span className="text-gray-400">App Version:</span>
                    <span className="text-gray-200 ml-2">{report.app_version}</span>
                  </div>
                </div>
              </div>

              {/* Connectivity */}
              <div className="bg-white/5 rounded-lg p-4 border border-white/5">
                <h4 className="text-sm font-bold text-white mb-3">Connectivity</h4>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-gray-400">Game Server</span>
                    <div className="flex items-center gap-2">
                      <StatusIcon status={report.connectivity.can_reach_game_server} />
                      <span className="text-xs text-gray-200">
                        {report.connectivity.can_reach_game_server ? 'Reachable' : 'Unreachable'}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-gray-400">Response Time</span>
                    <span className="text-xs text-gray-200">{report.connectivity.response_time_ms}ms</span>
                  </div>
                  {report.connectivity.game_server_error && (
                    <div className="mt-2 p-2 bg-red-500/10 border border-red-500/20 rounded text-xs text-red-300">
                      {report.connectivity.game_server_error}
                    </div>
                  )}
                </div>
              </div>

              {/* Installation */}
              <div className="bg-white/5 rounded-lg p-4 border border-white/5">
                <h4 className="text-sm font-bold text-white mb-3">Local Installation</h4>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-gray-400">Game</span>
                    <div className="flex items-center gap-2">
                      <StatusIcon status={report.local_installation.game_installed} />
                      <span className="text-xs text-gray-200">
                        {report.local_installation.game_installed 
                          ? `v${report.local_installation.current_version}` 
                          : 'Not installed'}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-gray-400">Java Runtime</span>
                    <StatusIcon status={report.local_installation.jre_installed} />
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-gray-400">Butler Tool</span>
                    <StatusIcon status={report.local_installation.butler_installed} />
                  </div>
                </div>
              </div>

              {/* Server Versions */}
              <div className="bg-white/5 rounded-lg p-4 border border-white/5">
                <h4 className="text-sm font-bold text-white mb-3">Server Versions</h4>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-gray-400">Latest Version</span>
                    <span className="text-xs text-gray-200">
                      {report.server_versions.latest_version || 'Not found'}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-gray-400">Versions Available</span>
                    <StatusIcon status={report.server_versions.found_versions} />
                  </div>
                  {report.server_versions.error && (
                    <div className="mt-2 p-2 bg-yellow-500/10 border border-yellow-500/20 rounded text-xs text-yellow-300">
                      {report.server_versions.error}
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Actions */}
        {report && (
          <div className="p-6 pt-0 flex gap-2">
            <button
              onClick={runDiagnostics}
              disabled={loading}
              className="flex-1 px-4 py-2 bg-white/5 hover:bg-white/10 rounded-lg border border-white/5 transition-colors disabled:opacity-50"
            >
              <span className="text-xs font-medium text-gray-300 flex items-center justify-center gap-2">
                <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
                Re-run
              </span>
            </button>
            <button
              onClick={copyReport}
              className="flex-1 px-4 py-2 bg-white/5 hover:bg-white/10 rounded-lg border border-white/5 transition-colors"
            >
              <span className="text-xs font-medium text-gray-300 flex items-center justify-center gap-2">
                <Copy size={14} />
                {copied ? 'Copied!' : 'Copy'}
              </span>
            </button>
            <button
              onClick={saveReport}
              className="flex-1 px-4 py-2 bg-[#FFA845]/20 hover:bg-[#FFA845]/30 rounded-lg border border-[#FFA845]/40 transition-colors"
            >
              <span className="text-xs font-medium text-white flex items-center justify-center gap-2">
                <Download size={14} />
                {saved ? 'Saved!' : 'Save'}
              </span>
            </button>
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};
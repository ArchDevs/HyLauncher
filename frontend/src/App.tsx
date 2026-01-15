import React, { useState, useEffect } from 'react';
import { Settings, FolderOpen, RefreshCw, Activity, ChevronDown, Edit3, Trash } from 'lucide-react';
import { motion } from 'framer-motion';
import BackgroundImage from './components/BackgroundImage';
import Titlebar from './components/Titlebar';
import { DeleteConfirmationModal } from './components/DeleteConfirmationModal';
import { ErrorModal } from './components/ErrorModal';
import { DiagnosticsModal } from './components/DiagnosticsModal';

import {
  DownloadAndLaunch,
  OpenFolder,
  GetVersions,
  GetNick,
  SetNick,
  DeleteGame,
  RunDiagnostics,
  SaveDiagnosticReport,
} from '../wailsjs/go/app/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

interface ProgressUpdate {
  stage: string;
  progress: number;
  message: string;
  currentFile: string;
  speed: string;
  downloaded: number;
  total: number;
}

interface AppError {
  type: string;
  message: string;
  technical: string;
  timestamp: string;
  stack?: string;
}

const App: React.FC = () => {
  const [username, setUsername] = useState("HyLauncher");
  const [isLoadingNick, setIsLoadingNick] = useState(true);
  const [current, setCurrent] = useState("");
  const [latest, setLatest] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [downloadProgress, setDownloadProgress] = useState(0);
  const [currentFile, setCurrentFile] = useState("");
  const [downloadSpeed, setDownloadSpeed] = useState("");
  const [downloaded, setDownloaded] = useState(0);
  const [total, setTotal] = useState(0);
  const [statusMessage, setStatusMessage] = useState("Ready to play");
  const [isDownloading, setIsDownloading] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [currentError, setCurrentError] = useState<AppError | null>(null);
  const [showDiagnostics, setShowDiagnostics] = useState(false);

  // Load nickname on startup
  useEffect(() => {
    const loadNickname = async () => {
      try {
        const nick = await GetNick();
        if (nick && nick.trim()) {
          setUsername(nick.trim());
        }
      } catch (err) {
        console.error("Failed to load nickname:", err);
        setStatusMessage("Warning: Could not load saved nickname");
      } finally {
        setIsLoadingNick(false);
      }
    };
    loadNickname();
  }, []);

  // Get game versions
  useEffect(() => {
    const fetchVersions = async () => {
      try {
        const [currentVersion, latestVersion] = await GetVersions();
        setCurrent(currentVersion);
        setLatest(latestVersion);
      } catch (err) {
        console.error("Failed to get versions:", err);
        setStatusMessage("Warning: Could not check game version");
      }
    };
    fetchVersions();
  }, []);

  // Listen for progress updates
  useEffect(() => {
    EventsOn('progress-update', (data: ProgressUpdate) => {
      setDownloadProgress(data.progress);
      setStatusMessage(data.message);
      setCurrentFile(data.currentFile);
      setDownloadSpeed(data.speed);
      setDownloaded(data.downloaded);
      setTotal(data.total);

      if (data.progress >= 100 && data.stage === 'launch') {
        setTimeout(() => {
          setIsDownloading(false);
          setDownloadProgress(0);
          setStatusMessage("Ready to play");
        }, 2000);
      }
    });

    // Listen for errors from backend
    EventsOn('error', (error: AppError) => {
      console.error('Backend error:', error);
      setCurrentError(error);
      setIsDownloading(false);
      setStatusMessage("Error occurred");
    });
  }, []);

  const saveNickname = async (newNick: string) => {
    const trimmed = newNick.trim();
    if (!trimmed || trimmed.length > 16) {
      setStatusMessage("Invalid nickname");
      return;
    }

    try {
      await SetNick(trimmed);
      setUsername(trimmed);
      setStatusMessage("Nickname saved");
      setTimeout(() => setStatusMessage("Ready to play"), 2000);
    } catch (err) {
      console.error("Failed to save nickname:", err);
      setStatusMessage("Failed to save nickname");
      setTimeout(() => setStatusMessage("Ready to play"), 3000);
    }
  };

  const handlePlay = async () => {
    const trimmed = username.trim();
    if (!trimmed) {
      setCurrentError({
        type: 'VALIDATION',
        message: 'Please enter a nickname before playing',
        technical: 'Empty nickname',
        timestamp: new Date().toISOString(),
      });
      return;
    }

    if (trimmed.length > 16) {
      setCurrentError({
        type: 'VALIDATION',
        message: 'Nickname is too long (maximum 16 characters)',
        technical: `Nickname length: ${trimmed.length}`,
        timestamp: new Date().toISOString(),
      });
      return;
    }

    setIsDownloading(true);
    setDownloadProgress(0);
    setStatusMessage("Starting...");
    
    try {
      await DownloadAndLaunch(trimmed);
    } catch (err) {
      console.error('Play error:', err);
      // Error will be emitted via 'error' event from backend
      setIsDownloading(false);
    }
  };

  const handleDeleteGame = async () => {
    setShowDeleteModal(false);
    setStatusMessage("Deleting game...");

    try {
      await DeleteGame();
      setStatusMessage("Game deleted successfully");
      setCurrent("");
      setTimeout(() => setStatusMessage("Ready to play"), 3000);
    } catch (err) {
      console.error("Delete error:", err);
      setStatusMessage("Failed to delete game");
      setTimeout(() => setStatusMessage("Ready to play"), 3000);
    }
  };

  const openGameFolder = async () => {
    try {
      await OpenFolder();
    } catch (err) {
      console.error('Open folder error:', err);
      setStatusMessage("Failed to open folder");
      setTimeout(() => setStatusMessage("Ready to play"), 3000);
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <div className="relative w-[1280px] h-[720px] bg-[#090909] text-white overflow-hidden font-sans select-none shadow-2xl rounded-[14px] border border-white/5">
      <BackgroundImage />
      <Titlebar />

      <main className="relative z-10 h-full p-10 flex flex-col justify-between pt-[60px]">
        {/* Top section */}
        <div className="flex justify-between items-start">
          <div className="flex flex-col gap-4">
            {/* Profile block */}
            <div className="w-[294px] h-[100px] bg-[#090909]/[0.55] backdrop-blur-xl rounded-[14px] border border-[#FFA845]/[0.10] p-4 flex flex-col justify-center gap-2">
              <div className="flex items-center justify-between">
                {isEditing ? (
                  <input
                    type="text"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    onBlur={() => {
                      setIsEditing(false);
                      saveNickname(username);
                    }}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        setIsEditing(false);
                        saveNickname(username);
                      }
                    }}
                    className="w-full bg-[#090909]/[0.55] border border-[#FFA845]/[0.10] rounded px-2 py-1 text-sm text-gray-200 focus:outline-none"
                    autoFocus
                    maxLength={16}
                  />
                ) : (
                  <>
                    <span className="text-sm font-medium text-gray-200">
                      {isLoadingNick ? "Loading..." : username}
                    </span>
                    <Edit3
                      size={14}
                      className="text-gray-400 cursor-pointer hover:text-white transition-colors"
                      onClick={() => setIsEditing(true)}
                    />
                  </>
                )}
              </div>

              <div className="flex items-center justify-between bg-[#090909]/[0.55] backdrop-blur-md rounded-lg px-3 py-2 border border-white/5 cursor-pointer hover:bg-white/5 transition-colors">
                <span className="text-xs text-gray-300">{current || "Not installed"}</span>
                <ChevronDown size={14} className="text-gray-400" />
              </div>
            </div>
          </div>

          {/* News section */}
          <div className="flex flex-col gap-4">
            {[1, 2, 3].map((i) => (
              <motion.div
                key={i}
                whileHover={{ x: -5, borderColor: 'rgba(255, 168, 69, 0.2)' }}
                className="w-[532px] h-[120px] bg-[#090909]/[0.55] backdrop-blur-xl rounded-[14px] border border-[#FFA845]/[0.10] p-4 flex gap-4 cursor-pointer"
              >
                <div className="flex-1">
                  <h3 className="text-sm font-bold text-gray-200 leading-snug">
                    Latest News: The update is almost here...
                  </h3>
                </div>
                <div className="w-[160px] h-full bg-[#090909]/[0.55] backdrop-blur-md rounded-lg border border-[#FFA845]/[0.10] flex items-center justify-center overflow-hidden">
                  <div className="text-[10px] text-[#FFA845]/[0.30] font-black uppercase tracking-widest">
                    Hytale
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* Bottom section */}
        <div className="w-full">
          <div className="flex items-end gap-8">
            {/* Left column - buttons + PLAY */}
            <div className="w-[294px] flex flex-col gap-3">
              <div className="flex gap-[10px]">
                <NavButton onClick={openGameFolder} icon={<FolderOpen size={20} />} />
                <NavButton 
                  onClick={() => setShowDiagnostics(true)} 
                  icon={<Activity size={20} />} 
                />
                <NavButton icon={<Settings size={20} />} />
                <NavButton
                  onClick={() => setShowDeleteModal(true)}
                  icon={<Trash size={20} />}
                />
              </div>

              <motion.button
                whileHover={{
                  scale: 1.01,
                  backgroundColor: 'rgba(9, 9, 9, 0.7)',
                  borderColor: 'rgba(255, 168, 69, 0.4)',
                }}
                whileTap={{ scale: 0.99 }}
                className="w-[294px] h-[94px] bg-[#090909]/[0.55] backdrop-blur-xl text-white font-black text-4xl tracking-tighter rounded-[14px] border border-[#FFA845]/[0.10] shadow-lg transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                onClick={handlePlay}
                disabled={isDownloading || isLoadingNick}
              >
                {isDownloading ? 'DOWNLOADING...' : 'PLAY'}
              </motion.button>
            </div>

            {/* Right column - progress */}
            <div className="flex-1 flex flex-col gap-4 pb-1">
              <div className="flex justify-between items-end">
                <div className="flex items-baseline gap-4">
                  <span className="text-5xl font-bold italic tracking-tighter">
                    {Math.round(downloadProgress)}%
                  </span>
                  <span className="text-[11px] text-gray-400 uppercase font-bold tracking-widest opacity-70">
                    {statusMessage}
                  </span>
                </div>

                <div className="text-[11px] text-gray-400 font-mono">
                  {downloadSpeed && total > 0
                    ? `${downloadSpeed} • ${formatBytes(downloaded)} / ${formatBytes(total)}`
                    : currentFile || 'Ready'}
                </div>
              </div>

              <div className="h-2 w-full bg-white/5 rounded-full overflow-hidden border border-white/5">
                <motion.div
                  animate={{ width: `${downloadProgress}%` }}
                  transition={{ duration: 0.3, ease: "easeOut" }}
                  className="h-full bg-white progress-glow"
                />
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Delete confirmation modal */}
      {showDeleteModal && (
        <DeleteConfirmationModal
          onConfirm={handleDeleteGame}
          onCancel={() => setShowDeleteModal(false)}
        />
      )}

      {/* Error modal */}
      {currentError && (
        <ErrorModal
          error={currentError}
          onClose={() => setCurrentError(null)}
        />
      )}

      {/* Diagnostics modal */}
      {showDiagnostics && (
        <DiagnosticsModal
          onClose={() => setShowDiagnostics(false)}
          onRunDiagnostics={RunDiagnostics}
          onSaveDiagnostics={SaveDiagnosticReport}
        />
      )}
    </div>
  );
};

interface NavButtonProps {
  icon: React.ReactNode;
  onClick?: () => void;
}

const NavButton: React.FC<NavButtonProps> = ({ icon, onClick }) => (
  <button
    onClick={onClick}
    className="w-[66px] h-[42px] flex items-center justify-center bg-[#090909]/[0.55] backdrop-blur-xl border border-[#FFA845]/[0.10] rounded-[14px] hover:bg-[#FFA845]/[0.05] hover:border-[#FFA845]/[0.30] transition-all cursor-pointer text-gray-400 hover:text-white"
  >
    {icon}
  </button>
);

export default App;
import React, { useState, useEffect } from 'react';
import { Settings, FolderOpen, RefreshCw, Gamepad2, ChevronDown, Edit3 } from 'lucide-react';
import { motion } from 'framer-motion';
import BackgroundImage from './components/BackgroundImage';
import Titlebar from './components/Titlebar';
import { DownloadAndLaunch } from '../wailsjs/go/main/App';
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

const App: React.FC = () => {
  // Global username state
  const [username, setUsername] = useState("Player01");
  const [isEditing, setIsEditing] = useState(false);
  const [downloadProgress, setDownloadProgress] = useState(0);
  const [currentFile, setCurrentFile] = useState("");
  const [downloadSpeed, setDownloadSpeed] = useState("");
  const [downloaded, setDownloaded] = useState(0);
  const [total, setTotal] = useState(0);
  const [statusMessage, setStatusMessage] = useState("Ready to play");
  const [isDownloading, setIsDownloading] = useState(false);

  useEffect(() => {
    // Listen for progress updates from backend
    EventsOn('progress-update', (data: ProgressUpdate) => {
      setDownloadProgress(data.progress);
      setStatusMessage(data.message);
      setCurrentFile(data.currentFile);
      setDownloadSpeed(data.speed);
      setDownloaded(data.downloaded);
      setTotal(data.total);

      // Reset downloading state when complete
      if (data.progress >= 100 && data.stage === 'launch') {
        setTimeout(() => {
          setIsDownloading(false);
          setDownloadProgress(0);
          setStatusMessage("Ready to play");
        }, 2000);
      }
    });
  }, []);

  const handlePlay = async () => {
    if (!username || !username.trim()) {
      return;
    }

    setIsDownloading(true);
    setDownloadProgress(0);
    setStatusMessage("Starting...");

    try {
      await DownloadAndLaunch(username.trim());
      console.log("Download started and game launched for:", username);
    } catch (err) {
      console.error(err);
      setStatusMessage("Error: " + err);
      setIsDownloading(false);
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
        
        {/* Top Section */}
        <div className="flex justify-between items-start">
          <div className="flex flex-col gap-4">
            {/* Profile Block */}
            <div className="w-[294px] h-[100px] bg-[#090909]/[0.55] backdrop-blur-xl rounded-[14px] border border-[#FFA845]/[0.10] p-4 flex flex-col justify-center gap-2">
              <div className="flex items-center justify-between">
                {isEditing ? (
                  <input
                    type="text"
                    value={username}
                    onChange={(e) => setUsername(e.target.value.slice(0, 16))}
                    onBlur={() => setIsEditing(false)}
                    onKeyDown={(e) => e.key === 'Enter' && setIsEditing(false)}
                    maxLength={16}
                    className="w-full bg-[#090909]/[0.55] border border-[#FFA845]/[0.10] rounded px-2 py-1 text-sm text-gray-200 focus:outline-none"
                    autoFocus
                    placeholder="Enter username (MAX: 16 symbols)"
                  />
                ) : (
                  <>
                    <span className="text-sm font-medium text-gray-200">{username}</span>
                    <Edit3
                      size={14}
                      className="text-gray-400 cursor-pointer hover:text-white transition-colors"
                      onClick={() => setIsEditing(true)}
                    />
                  </>
                )}
              </div>

              <div className="flex items-center justify-between bg-[#090909]/[0.55] backdrop-blur-md rounded-lg px-3 py-2 border border-white/5 cursor-pointer hover:bg-white/5 transition-colors">
                <span className="text-xs text-gray-300">Latest Version 1.0</span>
                <ChevronDown size={14} className="text-gray-400" />
              </div>
            </div>
          </div>

          {/* News Feed */}
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
                  <div className="text-[10px] text-[#FFA845]/[0.30] font-black uppercase tracking-widest">Hytale</div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* Bottom Section */}
        <div className="w-full">
          <div className="flex items-end gap-8">
            
            {/* Left Column */}
            <div className="w-[294px] flex flex-col gap-3">
              <div className="flex gap-[10px]">
                <NavButton icon={<Gamepad2 size={20}/>} />
                <NavButton icon={<FolderOpen size={20}/>} />
                <NavButton icon={<RefreshCw size={20}/>} />
                <NavButton icon={<Settings size={20}/>} />
              </div>

              <motion.button
                whileHover={{ scale: 1.01, backgroundColor: 'rgba(9, 9, 9, 0.7)', borderColor: 'rgba(255, 168, 69, 0.4)' }}
                whileTap={{ scale: 0.99 }}
                className="w-[294px] h-[94px] bg-[#090909]/[0.55] backdrop-blur-xl text-white font-black text-4xl tracking-tighter rounded-[14px] border border-[#FFA845]/[0.10] shadow-lg transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                onClick={handlePlay}
                disabled={isDownloading}
              >
                {isDownloading ? 'DOWNLOADING...' : 'PLAY'}
              </motion.button>
            </div>

            {/* Right Column */}
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
                  {downloadSpeed && total > 0 ? (
                    `${downloadSpeed} • ${formatBytes(downloaded)} / ${formatBytes(total)}`
                  ) : (
                    currentFile || 'Ready'
                  )}
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
    </div>
  );
};

const NavButton = ({ icon }: { icon: React.ReactNode }) => (
  <button className="w-[66px] h-[42px] flex items-center justify-center bg-[#090909]/[0.55] backdrop-blur-xl border border-[#FFA845]/[0.10] rounded-[14px] hover:bg-[#FFA845]/[0.05] hover:border-[#FFA845]/[0.30] transition-all cursor-pointer text-gray-400 hover:text-white">
    {icon}
  </button>
);

export default App;
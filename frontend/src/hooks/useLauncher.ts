import { useState, useEffect } from 'react';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { GetNick, GetLocalGameVersion, SetNick, DownloadAndLaunch } from '../../wailsjs/go/app/App';

export const useLauncher = () => {
  const [username, setUsername] = useState("HyLauncher");
  const [isLoadingNick, setIsLoadingNick] = useState(true);
  const [current, setCurrent] = useState(0);
  const [latest, setLatest] = useState("");
  const [statusMessage, setStatusMessage] = useState("Ready to play");
  const [isDownloading, setIsDownloading] = useState(false);
  const [downloadProgress, setDownloadProgress] = useState(0);
  const [downloadDetails, setDownloadDetails] = useState({ speed: '', currentFile: '', downloaded: 0, total: 0 });
  const [currentError, setCurrentError] = useState<any | null>(null);

  useEffect(() => {
    const init = async () => {
      try {
        const nick = await GetNick();
        if (nick?.trim()) setUsername(nick.trim());
        const curr = await GetLocalGameVersion("default")
        setCurrent(curr);
      } catch (err) {
        setStatusMessage("Warning: Connection issue");
      } finally {
        setIsLoadingNick(false);
      }
    };
    init();

    return EventsOn('progress-update', (data: any) => {
      setDownloadProgress(data.progress);
      setStatusMessage(data.message);
      setDownloadDetails({ 
        speed: data.speed, 
        currentFile: data.currentFile, 
        downloaded: data.downloaded, 
        total: data.total 
      });
      if (data.progress >= 100 && data.stage === 'launch') {
        setTimeout(() => { setIsDownloading(false); setDownloadProgress(0); setStatusMessage("Ready to play"); }, 2000);
      }
    });
  }, []);

  const handlePlay = async () => {
    if (!username.trim() || username.length > 16) {
      setCurrentError({ type: 'VALIDATION', message: 'Invalid Nickname', technical: 'Length check failed', timestamp: new Date().toISOString() });
      return;
    }
    setIsDownloading(true);
    try { await DownloadAndLaunch(username); } catch (err) { setIsDownloading(false); }
  };

  return { 
    username, setUsername, isLoadingNick, current, setCurrent, 
    statusMessage, setStatusMessage, isDownloading, setIsDownloading,
    downloadProgress, downloadDetails, currentError, setCurrentError, handlePlay 
  };
};
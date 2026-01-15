import React, { useState, useEffect } from 'react';
import BackgroundImage from './components/BackgroundImage';
import Titlebar from './components/Titlebar';
import { ProfileSection } from './components/ProfileCard';
import { UpdateOverlay } from './components/UpdateOverlay';
import { ControlSection } from './components/ControlSection';
import { DeleteConfirmationModal } from './components/DeleteConfirmationModal';
import { ErrorModal } from './components/ErrorModal';
import { DiagnosticsModal } from './components/DiagnosticsModal';

import { DownloadAndLaunch, OpenFolder, GetVersions, GetNick, SetNick, DeleteGame, RunDiagnostics, SaveDiagnosticReport, Update } from '../wailsjs/go/app/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

const App: React.FC = () => {
  const [username, setUsername] = useState<string>("HyLauncher");
  const [current, setCurrent] = useState<string>("");
  const [isEditing, setIsEditing] = useState<boolean>(false);
  const [progress, setProgress] = useState<number>(0);
  const [status, setStatus] = useState<string>("Ready to play");
  const [isDownloading, setIsDownloading] = useState<boolean>(false);
  
  // Новые стейты для деталей загрузки
  const [currentFile, setCurrentFile] = useState<string>("");
  const [downloadSpeed, setDownloadSpeed] = useState<string>("");
  const [downloaded, setDownloaded] = useState<number>(0);
  const [total, setTotal] = useState<number>(0);
  
  const [updateAsset, setUpdateAsset] = useState<any>(null);
  const [isUpdatingLauncher, setIsUpdatingLauncher] = useState<boolean>(false);
  const [updateStats, setUpdateStats] = useState({ d: 0, t: 0 });

  const [showDelete, setShowDelete] = useState<boolean>(false);
  const [showDiag, setShowDiag] = useState<boolean>(false);
  const [error, setError] = useState<any>(null);

  useEffect(() => {
    GetNick().then((n: string) => n && setUsername(n));
    
    GetVersions().then((v: any) => {
        if (Array.isArray(v)) setCurrent(v[0]);
        else setCurrent(v);
    });

    EventsOn('update:available', (asset: any) => setUpdateAsset(asset));
    EventsOn('update:progress', (d: number, t: number) => {
        setProgress((d/t)*100);
        setUpdateStats({ d, t });
    });

    EventsOn('progress-update', (data: any) => {
      setProgress(data.progress);
      setStatus(data.message);
      // Обновляем детальную инфу
      setCurrentFile(data.currentFile || "");
      setDownloadSpeed(data.speed || "");
      setDownloaded(data.downloaded || 0);
      setTotal(data.total || 0);

      if (data.progress >= 100 && data.stage === 'launch') {
        setTimeout(() => { 
            setIsDownloading(false); 
            setProgress(0); 
            setStatus("Ready to play"); 
            setDownloadSpeed("");
        }, 2000);
      }
    });
  }, []);

  return (
    <div className="relative w-screen h-screen max-w-[1280px] max-h-[720px] bg-[#090909] text-white overflow-hidden font-sans select-none rounded-[14px] border border-white/5 mx-auto">
      <BackgroundImage />
      <Titlebar />

      {isUpdatingLauncher && <UpdateOverlay progress={progress} downloaded={updateStats.d} total={updateStats.t} />}

      <main className="relative z-10 h-full p-10 flex flex-col justify-between pt-[60px]">
        <div className="flex justify-between items-start">
          <ProfileSection 
            username={username}
            currentVersion={current}
            isEditing={isEditing}
            onEditToggle={(val: boolean) => setIsEditing(val)}
            onUserChange={(val: string) => { SetNick(val); setUsername(val); }}
            updateAvailable={!!updateAsset}
            onUpdate={() => { setIsUpdatingLauncher(true); Update(); }}
          />

          <div className="w-[532px] h-[120px] bg-[#090909]/[0.55] backdrop-blur-xl rounded-[14px] border border-[#FFA845]/[0.10] p-4">
             <h3 className="text-sm font-bold text-gray-200">Latest News</h3>
          </div>
        </div>

        <ControlSection 
          onPlay={() => { setIsDownloading(true); DownloadAndLaunch(username); }}
          isDownloading={isDownloading}
          progress={progress}
          status={status}
          // Передаем новые пропсы
          speed={downloadSpeed}
          downloaded={downloaded}
          total={total}
          currentFile={currentFile}
          actions={{
            openFolder: OpenFolder,
            showDiagnostics: () => setShowDiag(true),
            showDelete: () => setShowDelete(true)
          }}
        />
      </main>

      {showDelete && <DeleteConfirmationModal onConfirm={() => { DeleteGame(); setShowDelete(false); }} onCancel={() => setShowDelete(false)} />}
      {showDiag && <DiagnosticsModal onClose={() => setShowDiag(false)} onRunDiagnostics={RunDiagnostics} onSaveDiagnostics={SaveDiagnosticReport} />}
      {error && <ErrorModal error={error} onClose={() => setError(null)} />}
    </div>
  );
};

export default App;
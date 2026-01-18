import React, { useState, useEffect } from 'react';
import BackgroundImage from './components/BackgroundImage';
import Titlebar from './components/Titlebar';
import { ProfileSection } from './components/ProfileCard';
import { UpdateOverlay } from './components/UpdateOverlay';
import { ControlSection } from './components/ControlSection';
import { DeleteConfirmationModal } from './components/DeleteConfirmationModal';
import { ErrorModal } from './components/ErrorModal';
import { DiagnosticsModal } from './components/DiagnosticsModal';

import { DownloadAndLaunch, OpenFolder, GetVersions, GetCurrentProfile, GetProfiles, SetCurrentProfile, AddProfile, UpdateProfile, DeleteProfile, DeleteGame, RunDiagnostics, SaveDiagnosticReport, Update } from '../wailsjs/go/app/App';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { config } from '../wailsjs/go/models';
import { NewsSection } from './components/NewsSection';
import { useWindowState } from './hooks/useWindowState';

// TODO FULL REFACTOR + Redesign

const App: React.FC = () => {
  const isMaximised = useWindowState();
  const [currentProfile, setCurrentProfile] = useState<config.Profile | null>(null);
  const [profiles, setProfiles] = useState<config.Profile[]>([]);
  const [current, setCurrent] = useState<string>("");
  const [progress, setProgress] = useState<number>(0);
  const [status, setStatus] = useState<string>("Ready to play");
  const [isDownloading, setIsDownloading] = useState<boolean>(false);

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

  const refreshProfiles = async () => {
    try {
      const [cp, ps] = await Promise.all([GetCurrentProfile(), GetProfiles()]);
      setCurrentProfile(cp);
      setProfiles(ps);
    } catch (err) {
      console.error("Failed to load profiles:", err);
    }
  };

  useEffect(() => {
    refreshProfiles();

    GetVersions().then((v: any) => {
      if (Array.isArray(v)) setCurrent(v[0]);
      else setCurrent(v);
    });

    EventsOn('update:available', (asset: any) => {
      console.log('Update available event received:', asset);
      setUpdateAsset(asset);
    });

    EventsOn('update:progress', (d: number, t: number) => {
      console.log(`Update progress: ${d}/${t} bytes`);
      const percentage = t > 0 ? (d / t) * 100 : 0;
      setProgress(percentage);
      setUpdateStats({ d, t });
    });

    EventsOn('progress-update', (data: any) => {
      setProgress(data.progress);
      setStatus(data.message);
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

  const handleUpdate = async () => {
    console.log('Update button clicked, starting update...');
    setIsUpdatingLauncher(true);
    setProgress(0);
    setUpdateStats({ d: 0, t: 0 });

    try {
      await Update();
      console.log('Update call completed');
    } catch (err) {
      console.error('Update failed:', err);
      setError({
        type: 'UPDATE_ERROR',
        message: 'Failed to update launcher',
        technical: err instanceof Error ? err.message : String(err),
        timestamp: new Date().toISOString()
      });
      setIsUpdatingLauncher(false);
    }
  };

  const handleProfileChange = async (id: string) => {
    await SetCurrentProfile(id);
    await refreshProfiles();
  };

  const handleProfileAdd = async (name: string) => {
    await AddProfile(name);
    await refreshProfiles();
  };

  const handleProfileUpdate = async (id: string, name: string) => {
    await UpdateProfile(id, name);
    await refreshProfiles();
  };

  const handleProfileDelete = async (id: string) => {
    await DeleteProfile(id);
    await refreshProfiles();
  };

  return (
    <div className={`relative w-full h-full bg-[#090909] text-white overflow-hidden font-sans select-none mx-auto ${isMaximised ? '' : 'border border-white/5'}`}>
      <BackgroundImage />
      <Titlebar isMaximised={isMaximised} />

      {isUpdatingLauncher && <UpdateOverlay progress={progress} downloaded={updateStats.d} total={updateStats.t} />}

      <main className="relative z-10 h-full p-10 flex flex-col justify-between pt-[60px]">
        <div className="flex justify-between items-start">
          <ProfileSection
            currentProfile={currentProfile}
            profiles={profiles}
            currentVersion={current}
            updateAvailable={!!updateAsset}
            onUpdate={handleUpdate}
            onProfileChange={handleProfileChange}
            onProfileAdd={handleProfileAdd}
            onProfileUpdate={handleProfileUpdate}
            onProfileDelete={handleProfileDelete}
          />

          <NewsSection />
        </div>

        <ControlSection
          onPlay={() => {
            if (currentProfile) {
              setIsDownloading(true);
              DownloadAndLaunch(currentProfile.name);
            }
          }}

          isDownloading={isDownloading}
          progress={progress}
          status={status}
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
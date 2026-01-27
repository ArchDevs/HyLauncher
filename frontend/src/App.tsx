import React, { useState, useEffect } from "react";
import BackgroundImage from "./components/BackgroundImage";
import Titlebar from "./components/Titlebar";
import { ProfileSection } from "./components/ProfileCard";
import { UpdateOverlay } from "./components/UpdateOverlay";
import { ControlSection } from "./components/ControlSection";
import { DeleteConfirmationModal } from "./components/DeleteConfirmationModal";
import { ErrorModal } from "./components/ErrorModal";

import {
  DownloadAndLaunch,
  OpenFolder,
  GetNick,
  SetNick,
  DeleteGame,
  Update,
  GetLocalGameVersion,
  GetLauncherVersion,
} from "../wailsjs/go/app/App";
import { EventsOn } from "../wailsjs/runtime/runtime";
import BannersHome from "./components/BannersHome";

// TODO FULL REFACTOR + Redesign

const App: React.FC = () => {
  const [username, setUsername] = useState<string>("HyLauncher");
  const [current, setCurrent] = useState<number>(0);
  const [launcherVersion, setLauncherVersion] = useState<string>("0.0.0");
  const [isEditing, setIsEditing] = useState<boolean>(false);
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

  useEffect(() => {
    GetNick().then((n: string) => n && setUsername(n));
    GetLocalGameVersion("default").then((curr: number) => setCurrent(curr));
    GetLauncherVersion().then((version: string) => setLauncherVersion(version));

    const offUpdateAvailable = EventsOn("update:available", (asset: any) => {
      console.log("Update available event received:", asset);
      setUpdateAsset(asset);
    });

    const offUpdateProgress = EventsOn(
      "update:progress",
      (d: number, t: number) => {
        const percentage = t > 0 ? (d / t) * 100 : 0;
        setProgress(percentage);
        setUpdateStats({ d, t });
      },
    );

    const offProgress = EventsOn("progress-update", (data: any) => {
      setProgress(data.progress ?? 0);
      setStatus(data.message ?? "");
      setCurrentFile(data.currentFile ?? "");
      setDownloadSpeed(data.speed ?? "");
      setDownloaded(data.downloaded ?? 0);
      setTotal(data.total ?? 0);

      if (data.stage === "launch") {
        setIsDownloading(false);
        setProgress(0);
        setStatus("Ready to play");
        setDownloadSpeed("");
      }

      if (data.stage === "idle") {
        setIsDownloading(false);
        setProgress(0);
        setStatus("Ready to play");
        setCurrentFile("");
        setDownloadSpeed("");
        setDownloaded(0);
        setTotal(0);
      }
    });

    return () => {
      offUpdateAvailable();
      offUpdateProgress();
      offProgress();
    };
  }, []);

  const handleUpdate = async () => {
    console.log("Update button clicked, starting update...");
    setIsUpdatingLauncher(true);
    setProgress(0);
    setUpdateStats({ d: 0, t: 0 });

    try {
      await Update();
      console.log("Update call completed");
    } catch (err) {
      console.error("Update failed:", err);
      setError({
        type: "UPDATE_ERROR",
        message: "Failed to update launcher",
        technical: err instanceof Error ? err.message : String(err),
        timestamp: new Date().toISOString(),
      });
      setIsUpdatingLauncher(false);
    }
  };

  return (
    <div className="relative w-screen h-screen max-w-[1280px] max-h-[720px] bg-[#090909] text-white overflow-hidden font-sans select-none rounded-[14px] border border-white/5 mx-auto">
      <BackgroundImage />
      <Titlebar />

      {isUpdatingLauncher && (
        <UpdateOverlay
          progress={progress}
          downloaded={updateStats.d}
          total={updateStats.t}
        />
      )}

      <main className="relative z-10 h-full p-10 flex flex-col justify-between pt-[60px]">
        <div className="flex justify-between items-start">
          <ProfileSection
            username={username}
            currentVersion={current}
            isEditing={isEditing}
            onEditToggle={(val: boolean) => setIsEditing(val)}
            onUserChange={(val: string) => {
              SetNick(val, "default");
              setUsername(val);
            }}
          />

          <BannersHome />
        </div>

        <ControlSection
          onPlay={() => {
            setIsDownloading(true);
            DownloadAndLaunch(username);
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
            showDelete: () => setShowDelete(true),
          }}
          updateAvailable={!!updateAsset}
          onUpdate={handleUpdate}
        />

        <div className="absolute right-[16px] bottom-[16px] text-[#FFFFFF]/[0.25] text-[14px] font-[Mazzard]">
          {launcherVersion}v
        </div>
      </main>

      {showDelete && (
        <DeleteConfirmationModal
          onConfirm={() => {
            DeleteGame("default");
            setShowDelete(false);
          }}
          onCancel={() => setShowDelete(false)}
        />
      )}
      {error && <ErrorModal error={error} onClose={() => setError(null)} />}
    </div>
  );
};

export default App;

import { useState, useEffect, useCallback } from "react";
import {
  DownloadAndLaunch,
  GetNick,
  SetNick as SetNickBackend,
  GetLocalGameVersion,
  SetLocalGameVersion,
  GetAvailableGameVersions,
  GetLauncherVersion,
  Update,
} from "../../wailsjs/go/app/App";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import { useTranslation } from "../i18n";

export const useLauncher = () => {
  const { t } = useTranslation();

  // Game
  const [username, setUsername] = useState<string>("HyLauncher");
  const [currentVersion, setCurrentVersion] = useState<number>(0);
  const [availableVersions, setAvailableVersions] = useState<number[]>([]);
  const [launcherVersion, setLauncherVersion] = useState<string>("0.0.0");
  const [isEditingUsername, setIsEditingUsername] = useState<boolean>(false);

  // Progress
  const [progress, setProgress] = useState<number>(0);
  const [status, setStatus] = useState<string>(t.control.status.readyToPlay);
  const [isDownloading, setIsDownloading] = useState<boolean>(false);

  // --- Download Details ---
  const [downloadDetails, setDownloadDetails] = useState({
    currentFile: "",
    speed: "",
    downloaded: 0,
    total: 0,
  });

  // Launcher update
  const [updateAsset, setUpdateAsset] = useState<any>(null);
  const [isUpdatingLauncher, setIsUpdatingLauncher] = useState<boolean>(false);
  const [updateStats, setUpdateStats] = useState({ d: 0, t: 0 });

  // UI
  const [showDeleteModal, setShowDeleteModal] = useState<boolean>(false);
  const [showDiagnostics, setShowDiagnostics] = useState<boolean>(false);
  const [error, setError] = useState<any>(null);

  useEffect(() => {
    // Initial data fetch
    GetNick().then((n: string) => n && setUsername(n));
    GetLocalGameVersion("default").then((curr: number) =>
      setCurrentVersion(curr),
    );
    GetAvailableGameVersions()
      .then((versions: number[]) => {
        // Sort descending so latest is first
        const sorted = [...versions].sort((a, b) => b - a);
        setAvailableVersions(sorted);
      })
      .catch(() => {
        // Non-fatal: just keep empty list; UI can handle gracefully
        setAvailableVersions([]);
      });
    GetLauncherVersion().then((version: string) => setLauncherVersion(version));

    // Listen for launcher updates
    const offUpdateAvailable = EventsOn("update:available", (asset: any) => {
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

    // Listen for game download progress
    const offProgress = EventsOn("progress-update", (data: any) => {
      setProgress(data.progress ?? 0);
      setStatus(data.message ?? "");
      setDownloadDetails({
        currentFile: data.currentFile ?? "",
        speed: data.speed ?? "",
        downloaded: data.downloaded ?? 0,
        total: data.total ?? 0,
      });

      if (data.stage === "idle") {
        setIsDownloading(false);
        setProgress(0);
        setStatus(t.control.status.readyToPlay);
        setDownloadDetails({
          currentFile: "",
          speed: "",
          downloaded: 0,
          total: 0,
        });
      }
    });

    return () => {
      offUpdateAvailable();
      offUpdateProgress();
      offProgress();
    };
  }, []);

  const handlePlay = useCallback(async () => {
    if (!username.trim()) {
      setError({ type: "VALIDATION", message: "Username cannot be empty" });
      return;
    }
    setIsDownloading(true);
    try {
      await DownloadAndLaunch(username);
    } catch (err) {
      setIsDownloading(false);
      setError({
        type: "LAUNCH_ERROR",
        message: "Failed to start game",
        technical: String(err),
      });
    }
  }, [username]);

  const handleUpdateLauncher = async () => {
    setIsUpdatingLauncher(true);
    setProgress(0);
    setUpdateStats({ d: 0, t: 0 });
    try {
      await Update();
    } catch (err) {
      setError({
        type: "UPDATE_ERROR",
        message: "Failed to update launcher",
        technical: err instanceof Error ? err.message : String(err),
        timestamp: new Date().toISOString(),
      });
      setIsUpdatingLauncher(false);
    }
  };

  const setNick = (val: string) => {
    SetNickBackend(val, "default");
    setUsername(val);
  };

  const setLocalGameVersion = async (version: number) => {
    try {
      await SetLocalGameVersion(version, "default");
      setCurrentVersion(version);
    } catch (err) {
      setError({
        type: "VERSION_ERROR",
        message: "Failed to set game version",
        technical: err instanceof Error ? err.message : String(err),
      });
    }
  };

  return {
    // State
    username,
    currentVersion,
    availableVersions,
    launcherVersion,
    isEditingUsername,
    setIsEditingUsername,
    progress,
    status,
    isDownloading,
    downloadDetails,
    updateAsset,
    isUpdatingLauncher,
    updateStats,
    showDeleteModal,
    setShowDeleteModal,
    showDiagnostics,
    setShowDiagnostics,
    error,
    setError,

    // Actions
    handlePlay,
    handleUpdateLauncher,
    setNick,
    setLocalGameVersion,
  };
};

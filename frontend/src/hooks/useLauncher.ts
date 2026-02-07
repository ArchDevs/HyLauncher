import { useState, useEffect, useCallback } from "react";
import {
  DownloadAndLaunch,
  GetNick,
  SetNick as SetNickBackend,
  GetInstanceInfo,
  GetAllGameVersions,
  GetLauncherVersion,
  Update,
} from "../../wailsjs/go/app/App";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import { useTranslation } from "../i18n";

export type ReleaseType = "release" | "pre-release";

export const useLauncher = () => {
  const { t } = useTranslation();

  // Game
  const [username, setUsername] = useState<string>("HyLauncher");
  const [currentVersion, setCurrentVersion] = useState<number>(0);
  const [selectedBranch, setSelectedBranch] = useState<ReleaseType>("pre-release");
  
  // Store versions for both branches
  const [allVersions, setAllVersions] = useState<{
    release: number[];
    preRelease: number[];
  }>({
    release: [],
    preRelease: [],
  });
  
  // Computed: current branch's available versions
  const availableVersions = selectedBranch === "release" 
    ? allVersions.release 
    : allVersions.preRelease;

  const [launcherVersion, setLauncherVersion] = useState<string>("0.0.0");
  const [isEditingUsername, setIsEditingUsername] = useState<boolean>(false);
  const [isLoadingVersions, setIsLoadingVersions] = useState<boolean>(true);

  // Progress
  const [progress, setProgress] = useState<number>(0);
  const [status, setStatus] = useState<string>(t.control.status.readyToPlay);
  const [isDownloading, setIsDownloading] = useState<boolean>(false);

  // Download Details
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

  // Load initial instance info from backend
  useEffect(() => {
    const loadInstanceInfo = async () => {
      try {
        const info = await GetInstanceInfo();
        console.log("[useLauncher] Loaded instance info:", info);
        
        setCurrentVersion(info.version || 0);
        setSelectedBranch((info.branch || "pre-release") as ReleaseType);
      } catch (err) {
        console.error("[useLauncher] Failed to load instance info:", err);
        setError({
          type: "INSTANCE_LOAD_ERROR",
          message: "Failed to load instance configuration",
          technical: err instanceof Error ? err.message : String(err),
        });
      }
    };

    loadInstanceInfo();
  }, []);

  // Fetch all versions on mount
  useEffect(() => {
    const fetchVersions = async () => {
      setIsLoadingVersions(true);
      try {
        const versions = await GetAllGameVersions();
        console.log("[useLauncher] Fetched versions:", versions);
        
        // Sort both arrays descending (latest first)
        const sortedRelease = [...(versions.release || [])].sort((a, b) => b - a);
        const sortedPreRelease = [...(versions.preRelease || [])].sort((a, b) => b - a);
        
        setAllVersions({
          release: sortedRelease,
          preRelease: sortedPreRelease,
        });
      } catch (err) {
        console.error("[useLauncher] Failed to fetch game versions:", err);
        setError({
          type: "VERSION_FETCH_ERROR",
          message: "Failed to fetch available game versions",
          technical: err instanceof Error ? err.message : String(err),
        });
        // Keep empty arrays on error
        setAllVersions({
          release: [],
          preRelease: [],
        });
      } finally {
        setIsLoadingVersions(false);
      }
    };

    fetchVersions();
  }, []);

  useEffect(() => {
    // Load username and launcher version
    GetNick().then((n: string) => {
      if (n) {
        console.log("[useLauncher] Loaded username:", n);
        setUsername(n);
      }
    }).catch(err => {
      console.error("[useLauncher] Failed to get username:", err);
    });

    GetLauncherVersion().then((version: string) => {
      console.log("[useLauncher] Launcher version:", version);
      setLauncherVersion(version);
    }).catch(err => {
      console.error("[useLauncher] Failed to get launcher version:", err);
    });

    // Listen for launcher updates
    const offUpdateAvailable = EventsOn("update:available", (asset: any) => {
      console.log("[useLauncher] Update available:", asset);
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
  }, [t.control.status.readyToPlay]);

  const handlePlay = useCallback(async () => {
    if (!username.trim()) {
      setError({ type: "VALIDATION", message: "Username cannot be empty" });
      return;
    }
    
    console.log("[useLauncher] Starting game with:", {
      username,
      version: currentVersion,
      branch: selectedBranch,
    });
    
    setIsDownloading(true);
    try {
      await DownloadAndLaunch(username);
      console.log("[useLauncher] Game launched successfully");
    } catch (err) {
      console.error("[useLauncher] Launch failed:", err);
      setIsDownloading(false);
      setError({
        type: "LAUNCH_ERROR",
        message: "Failed to start game",
        technical: String(err),
      });
    }
  }, [username, currentVersion, selectedBranch]);

  const handleUpdateLauncher = async () => {
    console.log("[useLauncher] Starting launcher update");
    setIsUpdatingLauncher(true);
    setProgress(0);
    setUpdateStats({ d: 0, t: 0 });
    try {
      await Update();
    } catch (err) {
      console.error("[useLauncher] Launcher update failed:", err);
      setError({
        type: "UPDATE_ERROR",
        message: "Failed to update launcher",
        technical: err instanceof Error ? err.message : String(err),
        timestamp: new Date().toISOString(),
      });
      setIsUpdatingLauncher(false);
    }
  };

  const setNick = useCallback((val: string) => {
    console.log("[useLauncher] Setting username:", val);
    SetNickBackend(val, "default");
    setUsername(val);
  }, []);

  // This is called by ProfileCard after backend confirms the change
  const setLocalGameVersion = useCallback((version: number) => {
    console.log("[useLauncher] Updating local version state:", version);
    setCurrentVersion(version);
  }, []);

  // This is called by ProfileCard after backend confirms the change
  const handleBranchChange = useCallback((branch: ReleaseType) => {
    console.log("[useLauncher] Updating local branch state:", branch);
    setSelectedBranch(branch);
  }, []);

  return {
    // State
    username,
    currentVersion,
    selectedBranch,
    availableVersions,
    allVersions,
    isLoadingVersions,
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
    handleBranchChange,
  };
};
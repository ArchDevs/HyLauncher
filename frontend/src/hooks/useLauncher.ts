import { useState, useEffect, useCallback } from "react";
import {
  DownloadAndLaunch,
  DownloadAndLaunchWithServer,
  GetNick,
  SetNick as SetNickBackend,
  GetInstanceInfo,
  GetReleaseVersions,
  GetPreReleaseVersions,
  GetLauncherVersion,
  Update,
  SetLocalGameVersion,
  UpdateInstanceBranch,
  GetAllNews,
  GetServers,
} from "../../wailsjs/go/app/App";
import { EventsOn, BrowserOpenURL } from "../../wailsjs/runtime/runtime";
import { useTranslation } from "../i18n";
import { model, service } from "../../wailsjs/go/models";

// Re-export server type
export type ServerWithFullUrls = service.ServerWithUrls;

export type ReleaseType = "release" | "pre-release";

export const useLauncher = () => {
  const { t } = useTranslation();

  const [username, setUsername] = useState<string>("HyLauncher");
  const [currentVersion, setCurrentVersion] = useState<string>("0");
  const [selectedBranch, setSelectedBranch] = useState<ReleaseType>("release");
  
  const [allVersions, setAllVersions] = useState<{
    release: string[];
    preRelease: string[];
  }>({
    release: [],
    preRelease: [],
  });
  
  const availableVersions = ["auto", ...(selectedBranch === "release" 
    ? allVersions.release 
    : allVersions.preRelease)];

  const [launcherVersion, setLauncherVersion] = useState<string>("0.0.0");
  const [isEditingUsername, setIsEditingUsername] = useState<boolean>(false);
  const [isLoadingVersions, setIsLoadingVersions] = useState<boolean>(true);
  const [allNews, setAllNews] = useState<any[]>([]);
  const [newsIndex, setNewsIndex] = useState<number>(0);
  
  // Servers state
  const [servers, setServers] = useState<service.ServerWithUrls[]>([]);
  const [isLoadingServers, setIsLoadingServers] = useState(true);

  useEffect(() => {
    if (allNews.length <= 1) return;
    
    const interval = setInterval(() => {
      setNewsIndex((prev) => (prev + 1) % allNews.length);
    }, 30000);
    
    return () => clearInterval(interval);
  }, [allNews.length]);

  const latestNews = allNews[newsIndex] || null;

  const [progress, setProgress] = useState<number>(0);
  const [status, setStatus] = useState<string>(t.control.status.readyToPlay);
  const [isDownloading, setIsDownloading] = useState<boolean>(false);

  const [downloadDetails, setDownloadDetails] = useState({
    currentFile: "",
    speed: "",
    downloaded: 0,
    total: 0,
  });

  const [updateAsset, setUpdateAsset] = useState<any>(null);
  const [isUpdatingLauncher, setIsUpdatingLauncher] = useState<boolean>(false);
  const [updateStats, setUpdateStats] = useState({ d: 0, t: 0 });

  const [showDeleteModal, setShowDeleteModal] = useState<boolean>(false);
  const [showDiagnostics, setShowDiagnostics] = useState<boolean>(false);
  const [error, setError] = useState<any>(null);

  // Initial load: get instance info then fetch versions for saved branch
  useEffect(() => {
    const loadInstanceInfo = async () => {
      try {
        const info = await GetInstanceInfo() as model.InstanceModel;
        const savedBranch = (info.Branch || "pre-release") as ReleaseType;
        
        setCurrentVersion(String(info.BuildVersion || "0"));
        setSelectedBranch(savedBranch);
        
        // Fetch versions for the saved branch
        setIsLoadingVersions(true);
        const response = savedBranch === "release" 
          ? await GetReleaseVersions()
          : await GetPreReleaseVersions();
        
        if (response.error) {
          throw new Error(response.error);
        }
        
        const sortedVersions = [...response.versions]
          .sort((a, b) => b - a)
          .map(v => String(v));
        
        setAllVersions(prev => ({
          release: savedBranch === "release" ? sortedVersions : prev.release,
          preRelease: savedBranch === "pre-release" ? sortedVersions : prev.preRelease,
        }));
      } catch (err) {
        setError({
          type: "INSTANCE_LOAD_ERROR",
          message: "Failed to load instance configuration",
          technical: err instanceof Error ? err.message : String(err),
        });
      } finally {
        setIsLoadingVersions(false);
      }
    };

    loadInstanceInfo();
  }, []);

  useEffect(() => {
    GetNick().then((n: string) => {
      if (n) setUsername(n);
    });

    GetLauncherVersion().then((version: string) => {
      setLauncherVersion(version);
    });

    GetAllNews().then((news: any[]) => {
      if (news && Array.isArray(news)) {
        setAllNews(news);
      }
    });
    
    // Fetch servers
    GetServers().then((data: service.ServerWithUrls[]) => {
      setServers(data);
      setIsLoadingServers(false);
    }).catch(() => {
      setIsLoadingServers(false);
    });

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

  const handlePlay = useCallback(async (serverIP?: string) => {
    if (!username.trim()) {
      setError({ type: "VALIDATION", message: "Username cannot be empty" });
      return;
    }
    
    setIsDownloading(true);
    try {
      if (serverIP) {
        await DownloadAndLaunchWithServer(username, serverIP);
      } else {
        await DownloadAndLaunch(username);
      }
    } catch (err) {
      setIsDownloading(false);
      setError({
        type: "LAUNCH_ERROR",
        message: "Failed to start game",
        technical: String(err),
      });
    }
  }, [username, currentVersion, selectedBranch]);

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

  const setNick = useCallback((val: string) => {
    SetNickBackend(val, "default");
    setUsername(val);
  }, []);

  const setLocalGameVersion = useCallback(async (version: string) => {
    setCurrentVersion(version);
    
    try {
      await SetLocalGameVersion(version, "default"); 
    } catch (err) {
      setError({ 
        type: "CONFIG_ERROR", 
        message: "Failed to save version", 
        technical: String(err) 
      });
    }
  }, [setError]);

  const handleBranchChange = useCallback(async (branch: ReleaseType) => {
    try {
      // 1. Save to backend FIRST (this must complete)
      await UpdateInstanceBranch(branch);
      
      // 2. Update UI state
      setSelectedBranch(branch);
      
      // 3. Fetch versions for the new branch
      setIsLoadingVersions(true);
      const response = branch === "release" 
        ? await GetReleaseVersions()
        : await GetPreReleaseVersions();
      
      if (response.error) {
        throw new Error(response.error);
      }
      
      const sortedVersions = [...response.versions]
        .sort((a, b) => b - a)
        .map(v => String(v));
        
      setAllVersions(prev => ({
        ...prev,
        [branch === "release" ? "release" : "preRelease"]: sortedVersions,
      }));
    } catch (err) {
      setError({ 
        type: "CONFIG_ERROR", 
        message: "Failed to save branch", 
        technical: String(err) 
      });
    } finally {
      setIsLoadingVersions(false);
    }
  }, [setError]);

  return {
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
    latestNews,
    onOpenNews: (url: string) => BrowserOpenURL(url),
    handlePlay,
    handleUpdateLauncher,
    setNick,
    setLocalGameVersion,
    handleBranchChange,
    // Servers
    servers,
    isLoadingServers,
  };
};
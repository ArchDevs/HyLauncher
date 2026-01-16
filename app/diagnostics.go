package app

import (
	"HyLauncher/internal/env"
	"HyLauncher/internal/pwr"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// DiagnosticReport contains diagnostic information
type DiagnosticReport struct {
	Timestamp         time.Time         `json:"timestamp"`
	AppVersion        string            `json:"app_version"`
	Platform          PlatformInfo      `json:"platform"`
	Connectivity      ConnectivityInfo  `json:"connectivity"`
	LocalInstallation InstallationInfo  `json:"local_installation"`
	ServerVersions    ServerVersionInfo `json:"server_versions"`
	DiskSpace         DiskSpaceInfo     `json:"disk_space"`
}

type PlatformInfo struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	GoVersion    string `json:"go_version"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
}

type ConnectivityInfo struct {
	CanReachGameServer bool   `json:"can_reach_game_server"`
	GameServerError    string `json:"game_server_error,omitempty"`
	ResponseTime       int64  `json:"response_time_ms"`
}

type InstallationInfo struct {
	GameInstalled   bool   `json:"game_installed"`
	CurrentVersion  string `json:"current_version"`
	InstallPath     string `json:"install_path"`
	JREInstalled    bool   `json:"jre_installed"`
	ButlerInstalled bool   `json:"butler_installed"`
}

type ServerVersionInfo struct {
	LatestVersion int      `json:"latest_version"`
	FoundVersions bool     `json:"found_versions"`
	CheckedURLs   []string `json:"checked_urls,omitempty"`
	Error         string   `json:"error,omitempty"`
}

type DiskSpaceInfo struct {
	InstallDirectory string `json:"install_directory"`
	Error            string `json:"error,omitempty"`
}

func (a *App) RunDiagnostics() (*DiagnosticReport, error) {
	report := &DiagnosticReport{
		Timestamp:  time.Now(),
		AppVersion: AppVersion,
	}

	// Platform info
	report.Platform = PlatformInfo{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
	}

	// Connectivity check
	report.Connectivity = checkConnectivity()

	// Local installation check
	report.LocalInstallation = checkLocalInstallation()

	// Server version check
	report.ServerVersions = checkServerVersions()

	// Disk space check
	report.DiskSpace = checkDiskSpace()

	return report, nil
}

func checkConnectivity() ConnectivityInfo {
	info := ConnectivityInfo{}

	start := time.Now()
	err := pwr.TestConnection()
	elapsed := time.Since(start).Milliseconds()

	info.ResponseTime = elapsed

	if err != nil {
		info.CanReachGameServer = false
		info.GameServerError = err.Error()
	} else {
		info.CanReachGameServer = true
	}

	return info
}

func checkLocalInstallation() InstallationInfo {
	info := InstallationInfo{
		InstallPath:    env.GetDefaultAppDir(),
		CurrentVersion: pwr.GetLocalVersion(),
	}

	// Check if game is installed
	gameClient := "HytaleClient"
	if runtime.GOOS == "windows" {
		gameClient += ".exe"
	}
	clientPath := filepath.Join(env.GetDefaultAppDir(), "release", "package", "game", "latest", "Client", gameClient)
	_, err := os.Stat(clientPath)
	info.GameInstalled = err == nil

	// Check if JRE is installed
	jreDir := filepath.Join(env.GetDefaultAppDir(), "release", "package", "jre", "latest")
	javaExec := filepath.Join(jreDir, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExec += ".exe"
	}
	_, err = os.Stat(javaExec)
	info.JREInstalled = err == nil

	// Check if Butler is installed
	butlerPath := filepath.Join(env.GetDefaultAppDir(), "tools", "butler", "butler")
	if runtime.GOOS == "windows" {
		butlerPath += ".exe"
	}
	_, err = os.Stat(butlerPath)
	info.ButlerInstalled = err == nil

	return info
}

func checkServerVersions() ServerVersionInfo {
	info := ServerVersionInfo{}

	result := pwr.FindLatestVersionWithDetails("release")

	info.LatestVersion = result.LatestVersion
	info.FoundVersions = result.LatestVersion > 0

	if len(result.CheckedURLs) > 0 {
		// Only include first few URLs to avoid huge reports
		maxURLs := 5
		if len(result.CheckedURLs) < maxURLs {
			maxURLs = len(result.CheckedURLs)
		}
		info.CheckedURLs = result.CheckedURLs[:maxURLs]
	}

	if result.Error != nil {
		info.Error = result.Error.Error()
	}

	return info
}

func checkDiskSpace() DiskSpaceInfo {
	info := DiskSpaceInfo{
		InstallDirectory: env.GetDefaultAppDir(),
	}

	// Note: Getting accurate disk space in a cross-platform way is complex
	// This is a simplified version
	stat, err := os.Stat(env.GetDefaultAppDir())
	if err != nil {
		info.Error = fmt.Sprintf("Cannot access install directory: %v", err)
		return info
	}

	if !stat.IsDir() {
		info.Error = "Install path is not a directory"
	}

	return info
}

// FormatDiagnosticReport formats the report as a human-readable string
func FormatDiagnosticReport(report *DiagnosticReport) string {
	output := fmt.Sprintf(`=== HyLauncher Diagnostic Report ===
Generated: %s
App Version: %s

--- Platform Information ---
OS: %s
Architecture: %s
CPUs: %d
Go Version: %s

--- Connectivity ---
Game Server Reachable: %v
Response Time: %dms
%s

--- Local Installation ---
Install Path: %s
Game Installed: %v
Current Version: %s
JRE Installed: %v
Butler Installed: %v

--- Server Versions ---
Latest Version Found: %d
Versions Available: %v
%s

--- Disk Space ---
Install Directory: %s
%s

`,
		report.Timestamp.Format("2006-01-02 15:04:05"),
		report.AppVersion,
		report.Platform.OS,
		report.Platform.Arch,
		report.Platform.NumCPU,
		report.Platform.GoVersion,
		report.Connectivity.CanReachGameServer,
		report.Connectivity.ResponseTime,
		formatConnectivityError(report.Connectivity),
		report.LocalInstallation.InstallPath,
		report.LocalInstallation.GameInstalled,
		report.LocalInstallation.CurrentVersion,
		report.LocalInstallation.JREInstalled,
		report.LocalInstallation.ButlerInstalled,
		report.ServerVersions.LatestVersion,
		report.ServerVersions.FoundVersions,
		formatServerError(report.ServerVersions),
		report.DiskSpace.InstallDirectory,
		formatDiskError(report.DiskSpace),
	)

	if len(report.ServerVersions.CheckedURLs) > 0 {
		output += "Sample URLs checked:\n"
		for _, url := range report.ServerVersions.CheckedURLs {
			output += fmt.Sprintf("  - %s\n", url)
		}
	}

	return output
}

func formatConnectivityError(info ConnectivityInfo) string {
	if info.GameServerError != "" {
		return fmt.Sprintf("Error: %s", info.GameServerError)
	}
	return ""
}

func formatServerError(info ServerVersionInfo) string {
	if info.Error != "" {
		return fmt.Sprintf("Error: %s", info.Error)
	}
	return ""
}

func formatDiskError(info DiskSpaceInfo) string {
	if info.Error != "" {
		return fmt.Sprintf("Error: %s", info.Error)
	}
	return "OK"
}

// SaveDiagnosticReport saves the diagnostic report to a file
func (a *App) SaveDiagnosticReport() (string, error) {
	report, err := a.RunDiagnostics()
	if err != nil {
		return "", err
	}

	diagnosticsDir := filepath.Join(env.GetDefaultAppDir(), "diagnostics")
	_ = os.MkdirAll(diagnosticsDir, 0755)

	filename := fmt.Sprintf("diagnostic_%s.txt", time.Now().Format("2006-01-02_15-04-05"))
	filepath := filepath.Join(diagnosticsDir, filename)

	content := FormatDiagnosticReport(report)

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return "", err
	}

	return filepath, nil
}

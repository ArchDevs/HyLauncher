package app

import (
	"HyLauncher/internal/patch"
	"HyLauncher/pkg/hyerrors"
)

// VersionsResponse is used to return versions without cyclic JSON references
type VersionsResponse struct {
	Versions []int  `json:"versions"`
	Error    string `json:"error,omitempty"`
}

func (a *App) DownloadAndLaunch(playerName string) error {
	return a.downloadAndLaunchInternal(playerName, "")
}

func (a *App) DownloadAndLaunchWithServer(playerName string, serverIP string) error {
	return a.downloadAndLaunchInternal(playerName, serverIP)
}

func (a *App) downloadAndLaunchInternal(playerName string, serverIP string) error {
	if err := a.validatePlayerName(playerName); err != nil {
		hyerrors.Report(hyerrors.Validation("provided invalid username"))
		return err
	}

	_ = a.SyncInstanceState()

	installedVersion, err := a.gameSvc.EnsureInstalled(a.ctx, a.instance, a.progress)
	if err != nil {
		appErr := hyerrors.WrapGame(err, "failed to install game").
			WithContext("branch", a.instance.Branch).
			WithContext("requestedVersion", a.instance.BuildVersion)
		hyerrors.Report(appErr)
		return appErr
	}

	if installedVersion != a.instance.BuildVersion {
		a.instance.BuildVersion = installedVersion
		if err := a.UpdateInstanceVersion(installedVersion); err != nil {
			_ = err
		}
	}

	if err := a.gameSvc.Launch(playerName, a.instance, serverIP); err != nil {
		appErr := hyerrors.GameCritical("failed to launch game").
			WithDetails(err.Error()).
			WithContext("player", playerName).
			WithContext("branch", a.instance.Branch).
			WithContext("version", a.instance.BuildVersion)
		hyerrors.Report(appErr)
		return appErr
	}

	return nil
}

func (a *App) GetAllGameVersions() (map[string]any, error) {
	release, prerelease, err := patch.ListAllVersionsBothBranches()
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"release":    release,
		"preRelease": prerelease,
	}, nil
}

func (a *App) GetReleaseVersions() VersionsResponse {
	release, err := patch.ListAllVersions("release")
	if err != nil {
		return VersionsResponse{Error: err.Error()}
	}
	return VersionsResponse{Versions: release}
}

func (a *App) GetPreReleaseVersions() VersionsResponse {
	release, err := patch.ListAllVersions("pre-release")
	if err != nil {
		return VersionsResponse{Error: err.Error()}
	}
	return VersionsResponse{Versions: release}
}

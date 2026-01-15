package updater

import (
	"encoding/json"
	"net/http"
	"runtime"
)

type UpdateInfo struct {
	Version string `json:"version"`
	Linux   struct {
		Amd64 Asset `json:"amd64"`
	} `json:"linux"`
	Windows struct {
		Amd64 Asset `json:"amd64"`
	} `json:"windows"`
}

type Asset struct {
	URL    string `json:"url"`
	Sha256 string `json:"sha256"`
}

func CheckUpdate(current string) (*Asset, string, error) {
	resp, err := http.Get(
		"https://github.com/ArchDevs/HyLauncher/releases/latest/download/version.json",
	)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, "", err
	}

	if info.Version == current {
		return nil, "", nil
	}

	if runtime.GOOS == "windows" {
		return &info.Windows.Amd64, info.Version, nil
	}
	return &info.Linux.Amd64, info.Version, nil
}

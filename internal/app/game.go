package app

import "HyLauncher/internal/patch"

type BranchVersions struct {
	Branch   string `json:"branch"`
	Versions []int  `json:"versions"`
}

type AllBranchVersions struct {
	Release    []int `json:"release"`
	PreRelease []int `json:"preRelease"`
}

func (a *App) GetAllGameVersions() (AllBranchVersions, error) {
	release, prerelease, err := patch.ListAllVersionsBothBranches()

	if err != nil {
		return AllBranchVersions{
			Release:    []int{},
			PreRelease: []int{},
		}, err
	}

	return AllBranchVersions{
		Release:    release,
		PreRelease: prerelease,
	}, nil
}

func (a *App) GetBranchVersions(branch string) ([]int, error) {
	versions, err := patch.ListAllVersions(branch)

	if err != nil {
		return []int{}, err
	}

	return versions, nil
}

func (a *App) GetLatestVersion(branch string) (int, error) {
	version, err := patch.FindLatestVersion(branch)

	if err != nil {
		return 0, err
	}

	return version, nil
}

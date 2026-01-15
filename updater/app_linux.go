//go:build linux

package updater

import "os"

func Apply(tmp string) error {
	exe, _ := os.Executable()
	return os.Rename(tmp, exe)
}

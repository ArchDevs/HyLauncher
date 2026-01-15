package game

import (
	"os"
	"os/exec"
	"runtime"
)

func isWayland() bool {
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	sessionType := os.Getenv("XDG_SESSION_TYPE")

	return waylandDisplay != "" || sessionType == "wayland"
}

func setSDLVideoDriver(cmd *exec.Cmd) {
	if runtime.GOOS == "linux" && isWayland() {
		env := os.Environ()
		env = append(env, "SDL_VIDEODRIVER=wayland")
		cmd.Env = env
	}
}

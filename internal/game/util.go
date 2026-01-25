package game

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/google/uuid"
)

func OfflineUUID(nick string) uuid.UUID {
	data := []byte("OfflinePlayer:" + strings.TrimSpace(nick))
	return uuid.NewMD5(uuid.Nil, data)
}

// Wayland
func SetSDLVideoDriver(cmd *exec.Cmd) {
	if runtime.GOOS == "linux" && isWayland() {
		env := os.Environ()
		env = append(env, "SDL_VIDEODRIVER=wayland")
		cmd.Env = env
	}
}

func isWayland() bool {
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	sessionType := os.Getenv("XDG_SESSION_TYPE")

	return waylandDisplay != "" || sessionType == "wayland"
}


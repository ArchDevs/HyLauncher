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
		withEnv(cmd, "SDL_VIDEODRIVER=wayland")
	}
}

func isWayland() bool {
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	sessionType := os.Getenv("XDG_SESSION_TYPE")

	return waylandDisplay != "" || sessionType == "wayland"
}

func withEnv(cmd *exec.Cmd, kv ...string) {
	baseEnv := os.Environ()
	if len(cmd.Env) > 0 {
		baseEnv = cmd.Env
	}

	envMap := make(map[string]string, len(baseEnv)+len(kv))
	for _, e := range baseEnv {
		if k, v, ok := strings.Cut(e, "="); ok {
			envMap[k] = v
		}
	}

	for _, e := range kv {
		if k, v, ok := strings.Cut(e, "="); ok {
			envMap[k] = v
		}
	}

	cmd.Env = cmd.Env[:0]
	for k, v := range envMap {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
}

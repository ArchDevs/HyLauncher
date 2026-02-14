package sysinfo

import (
	"fmt"
	"os"
	"runtime"

	"github.com/jaypipes/ghw"
)

// SystemInfo holds detailed system information
type SystemInfo struct {
	OS       OSInfo        `json:"os"`
	CPU      CPUInfo       `json:"cpu"`
	Memory   MemoryInfo    `json:"memory"`
	GPU      []GPUInfo     `json:"gpu"`
	Displays []DisplayInfo `json:"displays"`
}

type OSInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Kernel    string `json:"kernel"`
	Arch      string `json:"arch"`
	GoVersion string `json:"go_version"`
}

type CPUInfo struct {
	Model   string `json:"model"`
	Cores   int    `json:"cores"`
	Threads int    `json:"threads"`
}

type MemoryInfo struct {
	Total string `json:"total"`
}

type GPUInfo struct {
	Model  string `json:"model"`
	Vendor string `json:"vendor"`
}

type DisplayInfo struct {
	Resolution string `json:"resolution"`
}

func GetSystemInfo() *SystemInfo {
	info := &SystemInfo{
		OS: OSInfo{
			Arch:      runtime.GOOS + "/" + runtime.GOARCH,
			GoVersion: runtime.Version(),
		},
	}

	if cpu, err := ghw.CPU(); err == nil && len(cpu.Processors) > 0 {
		info.CPU.Model = cpu.Processors[0].Model
		info.CPU.Cores = int(cpu.TotalCores)
		info.CPU.Threads = int(cpu.TotalThreads)
	}

	if mem, err := ghw.Memory(); err == nil {
		info.Memory.Total = formatBytes(int64(mem.TotalUsableBytes))
	}

	if gpu, err := ghw.GPU(); err == nil {
		for _, card := range gpu.GraphicsCards {
			g := GPUInfo{}
			if card.DeviceInfo != nil {
				if card.DeviceInfo.Product != nil {
					g.Model = card.DeviceInfo.Product.Name
				}
				if card.DeviceInfo.Vendor != nil {
					g.Vendor = card.DeviceInfo.Vendor.Name
				}
			}
			if g.Model != "" {
				info.GPU = append(info.GPU, g)
			}
		}
	}

	info.Displays = append(info.Displays, DisplayInfo{
		Resolution: os.Getenv("DISPLAY"),
	})

	return info
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

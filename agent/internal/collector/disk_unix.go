//go:build !windows

package collector

import (
	"os"
	"strings"
	"syscall"

	"github.com/pingan/hydra-agent/internal/model"
)

func collectDiskMetrics(batch *model.MetricsBatch) {
	mounts, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(mounts), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		device := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]

		if isPseudoFS(fsType) || strings.HasPrefix(device, "none") {
			continue
		}

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mountPoint, &stat); err != nil {
			continue
		}

		totalBytes := stat.Blocks * uint64(stat.Bsize)
		availBytes := stat.Bavail * uint64(stat.Bsize)
		usedBytes := totalBytes - availBytes

		labels := model.Labels{
			"mount":     mountPoint,
			"device":    device,
			"fstype":    fsType,
			"component": "disk",
		}

		batch.Add("disk_total_bytes", float64(totalBytes), labels)
		batch.Add("disk_used_bytes", float64(usedBytes), labels)
		if totalBytes > 0 {
			batch.Add("disk_usage_percent", float64(usedBytes)/float64(totalBytes)*100.0, labels)
		}
	}
}

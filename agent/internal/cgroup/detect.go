package cgroup

import "os"

// Version represents the detected cgroup version.
type Version int

const (
	V1 Version = 1
	V2 Version = 2
)

// Detect probes the filesystem to determine which cgroup version is in use.
// cgroup v2 has a unified hierarchy mounted at /sys/fs/cgroup with a
// cgroup.controllers file at the root. v1 uses multiple subsystems each in
// their own directory.
func Detect() Version {
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err == nil {
		return V2
	}
	return V1
}

// IsV2 is a convenience check.
func IsV2() bool {
	return Detect() == V2
}

// CPUStatPath returns the path to the CPU usage statistics file.
func CPUStatPath() string {
	if IsV2() {
		return "/sys/fs/cgroup/cpu.stat"
	}
	return "/proc/stat"
}

// MemoryCurrentPath returns the path to the current memory usage file.
func MemoryCurrentPath() string {
	if IsV2() {
		return "/sys/fs/cgroup/memory.current"
	}
	return "/proc/meminfo"
}

// MemoryStatPath returns the path to memory event counters (OOM kills, etc.).
func MemoryStatPath() string {
	if IsV2() {
		return "/sys/fs/cgroup/memory.events"
	}
	return ""
}

// MemoryMaxPath returns the path to the memory hard limit file (v2 only).
func MemoryMaxPath() string {
	if IsV2() {
		return "/sys/fs/cgroup/memory.max"
	}
	return ""
}

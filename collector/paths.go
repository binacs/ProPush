package collector

import (
	"path/filepath"
)

func procFilePath(name string) string {
	return filepath.Join("/proc", name)
}

func sysFilePath(name string) string {
	return filepath.Join("/sys", name)
}

func rootfsFilePath(name string) string {
	return filepath.Join("/", name)
}

func rootfsStripPrefix(path string) string {
	return path
}

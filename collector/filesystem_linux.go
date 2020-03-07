package collector

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/unix"
)

var (
	filesystemLabelNames = []string{"device", "mountpoint", "fstype"}
)

type filesystemCollector struct {
	usageDesc *prometheus.Desc
}

type filesystemLabels struct {
	device, mountPoint, fsType, options string
}

type filesystemStats struct {
	labels            filesystemLabels
	size, free, avail float64
	files, filesFree  float64
	ro, deviceError   float64
}

// NewFilesystemCollector returns a new Collector exposing filesystems stats.
func NewFilesystemCollector() (Collector, error) {
	subsystem := "disk"

	usageDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "usage"),
		"Filesystem usage.",
		filesystemLabelNames, nil,
	)

	return &filesystemCollector{
		usageDesc: usageDesc,
	}, nil
}

func (c *filesystemCollector) Update(ch chan<- prometheus.Metric) error {
	stats, err := c.GetStats()
	if err != nil {
		return err
	}
	// Make sure we expose a metric once, even if there are multiple mounts
	seen := map[filesystemLabels]bool{}
	for _, s := range stats {
		if seen[s.labels] {
			continue
		}
		seen[s.labels] = true

		if s.deviceError > 0 {
			continue
		}
		// only use "/"
		if s.labels.mountPoint != "/" {
			continue
		}
		if s.labels.device[0] != '/' {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.usageDesc, prometheus.GaugeValue,
			100-s.avail/s.size*100, s.labels.device, s.labels.mountPoint, s.labels.fsType,
		)
	}
	return nil
}

var mountTimeout = 5 * time.Second // 5s
var stuckMounts = make(map[string]struct{})
var stuckMountsMtx = &sync.Mutex{}

// GetStats returns filesystem stats.
func (c *filesystemCollector) GetStats() ([]filesystemStats, error) {
	mps, err := mountPointDetails()
	if err != nil {
		return nil, err
	}
	stats := []filesystemStats{}
	for _, labels := range mps {
		stuckMountsMtx.Lock()
		if _, ok := stuckMounts[labels.mountPoint]; ok {
			stats = append(stats, filesystemStats{
				labels:      labels,
				deviceError: 1,
			})
			stuckMountsMtx.Unlock()
			continue
		}
		stuckMountsMtx.Unlock()

		// The success channel is used do tell the "watcher" that the stat
		// finished successfully. The channel is closed on success.
		success := make(chan struct{})
		go stuckMountWatcher(labels.mountPoint, success)

		buf := new(unix.Statfs_t)
		err = unix.Statfs(rootfsFilePath(labels.mountPoint), buf)
		stuckMountsMtx.Lock()
		close(success)
		// If the mount has been marked as stuck, unmark it and log it's recovery.
		if _, ok := stuckMounts[labels.mountPoint]; ok {
			delete(stuckMounts, labels.mountPoint)
		}
		stuckMountsMtx.Unlock()

		if err != nil {
			stats = append(stats, filesystemStats{
				labels:      labels,
				deviceError: 1,
			})

			continue
		}

		var ro float64
		for _, option := range strings.Split(labels.options, ",") {
			if option == "ro" {
				ro = 1
				break
			}
		}

		stats = append(stats, filesystemStats{
			labels:    labels,
			size:      float64(buf.Blocks) * float64(buf.Bsize),
			free:      float64(buf.Bfree) * float64(buf.Bsize),
			avail:     float64(buf.Bavail) * float64(buf.Bsize),
			files:     float64(buf.Files),
			filesFree: float64(buf.Ffree),
			ro:        ro,
		})
	}
	return stats, nil
}

// stuckMountWatcher listens on the given success channel and if the channel closes
// then the watcher does nothing. If instead the timeout is reached, the
// mount point that is being watched is marked as stuck.
func stuckMountWatcher(mountPoint string, success chan struct{}) {
	select {
	case <-success:
		// Success
	case <-time.After(mountTimeout):
		// Timed out, mark mount as stuck
		stuckMountsMtx.Lock()
		select {
		case <-success:
			// Success came in just after the timeout was reached, don't label the mount as stuck
		default:
			stuckMounts[mountPoint] = struct{}{}
		}
		stuckMountsMtx.Unlock()
	}
}

func mountPointDetails() ([]filesystemLabels, error) {
	file, err := os.Open(procFilePath("1/mounts"))
	if os.IsNotExist(err) {
		// Fallback to `/proc/mounts` if `/proc/1/mounts` is missing due hidepid.
		file, err = os.Open(procFilePath("mounts"))
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseFilesystemLabels(file)
}

func parseFilesystemLabels(r io.Reader) ([]filesystemLabels, error) {
	var filesystems []filesystemLabels

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())

		if len(parts) < 4 {
			return nil, fmt.Errorf("malformed mount point information: %q", scanner.Text())
		}

		// Ensure we handle the translation of \040 and \011
		// as per fstab(5).
		parts[1] = strings.Replace(parts[1], "\\040", " ", -1)
		parts[1] = strings.Replace(parts[1], "\\011", "\t", -1)

		filesystems = append(filesystems, filesystemLabels{
			device:     parts[0],
			mountPoint: rootfsStripPrefix(parts[1]),
			fsType:     parts[2],
			options:    parts[3],
		})
	}

	return filesystems, scanner.Err()
}

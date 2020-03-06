package collector

import (
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
)

const (
	cpuCollectorSubsystem = "cpu"
)

var (
	nodeCPUSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, cpuCollectorSubsystem, "usage"),
		"Usage of the cpus in 100ms.",
		[]string{"cpu", "intervel"}, nil,
	)
)

type cpuCollector struct {
	fs  procfs.FS
	cpu *prometheus.Desc
}

func NewCPUCollector() (Collector, error) {
	fs, err := procfs.NewFS("/proc") // proc: default path
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}
	return &cpuCollector{
		fs:  fs,
		cpu: nodeCPUSecondsDesc,
	}, nil
}

func (c *cpuCollector) Update(ch chan<- prometheus.Metric) error {
	if err := c.updateStat(ch); err != nil {
		return err
	}
	return nil
}

func (c *cpuCollector) updateStat(ch chan<- prometheus.Metric) error {
	statsS, err := c.fs.Stat()
	if err != nil {
		return err
	}
	time.Sleep(100000000)
	statsE, err := c.fs.Stat()
	if err != nil {
		return err
	}
	if len(statsE.CPU) < len(statsS.CPU) {
		return nil
	}

	for cpuID, cpuStatS := range statsS.CPU {
		cpuStatE := statsE.CPU[cpuID]
		cpuNum := strconv.Itoa(cpuID)
		totalS := cpuStatS.User + cpuStatS.Nice + cpuStatS.System + cpuStatS.Idle + cpuStatS.Iowait + cpuStatS.IRQ + cpuStatS.SoftIRQ
		totalE := cpuStatE.User + cpuStatE.Nice + cpuStatE.System + cpuStatE.Idle + cpuStatE.Iowait + cpuStatE.IRQ + cpuStatE.SoftIRQ
		var usage float64
		if cpuStatE.Idle == cpuStatS.Idle {
			usage = 0
		} else {
			usage = 100 - (cpuStatE.Idle-cpuStatS.Idle)/(totalE-totalS)*100
		}
		ch <- prometheus.MustNewConstMetric(c.cpu, prometheus.CounterValue, usage, cpuNum, "100ms")
	}

	return nil
}

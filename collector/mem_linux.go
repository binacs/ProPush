package collector

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	memInfoSubsystem = "memory"
)

type meminfoCollector struct {
}

// NewMeminfoCollector returns a new Collector exposing memory stats.
func NewMeminfoCollector() (Collector, error) {
	return &meminfoCollector{}, nil
}

// Update calls (*meminfoCollector).getMemInfo to get the platform specific
// memory metrics.
func (c *meminfoCollector) Update(ch chan<- prometheus.Metric) error {
	memInfo, err := c.getMemInfo()
	if err != nil {
		return fmt.Errorf("couldn't get meminfo: %s", err)
	}

	usage := 100 - (memInfo["MemFree_bytes"]+memInfo["Cached_bytes"]+memInfo["Buffers_bytes"])/memInfo["MemTotal_bytes"]*100

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, memInfoSubsystem, "usage"),
			fmt.Sprintf("Memory usage."),
			nil, nil,
		),
		prometheus.GaugeValue, usage,
	)

	return nil
}

func (c *meminfoCollector) getMemInfo() (map[string]float64, error) {
	file, err := os.Open(procFilePath("meminfo"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseMemInfo(file)
}

func parseMemInfo(r io.Reader) (map[string]float64, error) {
	var (
		memInfo = map[string]float64{}
		scanner = bufio.NewScanner(r)
		re      = regexp.MustCompile(`\((.*)\)`)
	)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		fv, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value in meminfo: %s", err)
		}
		key := parts[0][:len(parts[0])-1] // remove trailing : from key
		// Active(anon) -> Active_anon
		key = re.ReplaceAllString(key, "_${1}")
		switch len(parts) {
		case 2: // no unit
		case 3: // has unit, we presume kB
			fv *= 1024
			key = key + "_bytes"
		default:
			return nil, fmt.Errorf("invalid line in meminfo: %s", line)
		}
		memInfo[key] = fv
	}

	return memInfo, scanner.Err()
}

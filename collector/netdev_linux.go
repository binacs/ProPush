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

type netDevCollector struct {
	subsystem   string
	metricDescs map[string]*prometheus.Desc
}

// NewNetDevCollector returns a new Collector exposing network device stats.
func NewNetDevCollector() (Collector, error) {
	return &netDevCollector{
		subsystem:   "network",
		metricDescs: map[string]*prometheus.Desc{},
	}, nil
}

func (c *netDevCollector) Update(ch chan<- prometheus.Metric) error {
	netDev, err := getNetDevStats(nil, nil)
	if err != nil {
		return fmt.Errorf("couldn't get netstats: %s", err)
	}
	var res float64
	for _, devStats := range netDev {
		for key, value := range devStats {
			if key != "receive_bytes" && key != "transmit_bytes" {
				continue
			}
			desc, ok := c.metricDescs[key]
			if !ok {
				desc = prometheus.NewDesc(
					prometheus.BuildFQName(namespace, c.subsystem, key+"_total"),
					fmt.Sprintf("Network device statistic %s.", key),
					[]string{"device"},
					nil,
				)
				c.metricDescs[key] = desc
			}
			v, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid value %s in netstats: %s", value, err)
			}
			res += v
			//ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, v, dev)
		}
	}
	// Bytes => MB
	res = res / 1024 / 1024
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, c.subsystem, "netio"),
			"Network I/O (MB).",
			nil,
			nil,
		), prometheus.GaugeValue, res)

	return nil
}

var (
	procNetDevInterfaceRE = regexp.MustCompile(`^(.+): *(.+)$`)
	procNetDevFieldSep    = regexp.MustCompile(` +`)
)

func getNetDevStats(ignore *regexp.Regexp, accept *regexp.Regexp) (map[string]map[string]string, error) {
	file, err := os.Open(procFilePath("net/dev"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseNetDevStats(file, ignore, accept)
}

func parseNetDevStats(r io.Reader, ignore *regexp.Regexp, accept *regexp.Regexp) (map[string]map[string]string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Scan() // skip first header
	scanner.Scan()
	parts := strings.Split(scanner.Text(), "|")
	if len(parts) != 3 { // interface + receive + transmit
		return nil, fmt.Errorf("invalid header line in net/dev: %s",
			scanner.Text())
	}

	receiveHeader := strings.Fields(parts[1])
	transmitHeader := strings.Fields(parts[2])
	headerLength := len(receiveHeader) + len(transmitHeader)

	netDev := map[string]map[string]string{}
	for scanner.Scan() {
		line := strings.TrimLeft(scanner.Text(), " ")
		parts := procNetDevInterfaceRE.FindStringSubmatch(line)
		if len(parts) != 3 {
			return nil, fmt.Errorf("couldn't get interface name, invalid line in net/dev: %q", line)
		}

		dev := parts[1]
		if ignore != nil && ignore.MatchString(dev) {
			continue
		}
		if accept != nil && !accept.MatchString(dev) {
			continue
		}

		values := procNetDevFieldSep.Split(strings.TrimLeft(parts[2], " "), -1)
		if len(values) != headerLength {
			return nil, fmt.Errorf("couldn't get values, invalid line in net/dev: %q", parts[2])
		}

		netDev[dev] = map[string]string{}
		for i := 0; i < len(receiveHeader); i++ {
			netDev[dev]["receive_"+receiveHeader[i]] = values[i]
		}

		for i := 0; i < len(transmitHeader); i++ {
			netDev[dev]["transmit_"+transmitHeader[i]] = values[i+len(receiveHeader)]
		}
	}
	return netDev, scanner.Err()
}

package collector

import (
	"errors"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "propush"

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"node_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"node_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

var (
	factories        = make(map[string]func() (Collector, error))
	collectorState   = make(map[string]bool)
	forcedCollectors = map[string]bool{} // collectors which have been explicitly enabled or disabled
)

func registerCollector(collector string, factory func() (Collector, error)) {
	factories[collector] = factory
}

func init() {
	collectorState["cpu"] = true
	collectorState["mem"] = true
	collectorState["disk"] = true
	collectorState["netio"] = true

	registerCollector("cpu", NewCPUCollector)
	registerCollector("mem", NewMeminfoCollector)
	registerCollector("disk", NewFilesystemCollector)
	registerCollector("netio", NewNetDevCollector)
}

type NodeCollector struct {
	Collectors map[string]Collector
}

func NewNodeCollector() *NodeCollector {
	collectors := make(map[string]Collector)
	for key, enabled := range collectorState {
		if enabled {
			collector, err := factories[key]()
			if err != nil {
				return nil
			}
			collectors[key] = collector
		}
	}
	return &NodeCollector{Collectors: collectors}
}

func (n NodeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

func (n NodeCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(n.Collectors))
	for name, c := range n.Collectors {
		go func(name string, c Collector) {
			execute(name, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		if IsNoDataError(err) {
			//level.Debug(logger).Log("msg", "collector returned no data", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		} else {
			//level.Error(logger).Log("msg", "collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		}
		success = 0
	} else {
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

type Collector interface {
	Update(ch chan<- prometheus.Metric) error
}

type typedDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

func (d *typedDesc) mustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(d.desc, d.valueType, value, labels...)
}

var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return err == ErrNoData
}

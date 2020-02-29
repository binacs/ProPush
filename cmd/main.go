package main

import (
	"bytes"
	"fmt"
	_ "net/http/pprof"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"

	"github.com/BinacsLee/ProPush/collector"
	"github.com/prometheus/client_golang/prometheus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	r := prometheus.NewRegistry()
	nc, err := collector.NewNodeCollector(logger)
	if err != nil {
		fmt.Println("couldn't create collector: ", err)
	}
	fmt.Println("nc = ", nc.Collectors)
	if err := r.Register(nc); err != nil {
		fmt.Println("couldn't register node collector: ", err)
	}

	//for {
	mfs, err := r.Gather()
	if err != nil {
		fmt.Println("gather err = ", err)
	}
	buf := &bytes.Buffer{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtText)
	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetName() == "job" {
					fmt.Println("pushed metric ", mf.GetName(), m, "already contains a job label")
				}
			}
		}
		if err := enc.Encode(mf); err != nil {
			fmt.Println("encode error", err)
		}
	}
	fmt.Println(buf.String())
	fmt.Println()
	//time.Sleep(1000000000)
	//}
}

package collector

import (
	"bytes"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

type Instance struct {
	R *prometheus.Registry
	C *NodeCollector
}

var instance *Instance

func GetInstance() *Instance {
	if instance == nil {
		instance = &Instance{
			R: prometheus.NewRegistry(),
			C: NewNodeCollector(),
		}
		if instance.R == nil || instance.C == nil || instance.R.Register(instance.C) != nil {
			instance = nil
		}
	}
	return instance
}

func (ins *Instance) GetMetrics() string {
	if ins == nil {
		return ""
	}
	mfs, err := ins.R.Gather()
	if err != nil {
		fmt.Println("gather err = ", err)
	}
	buf := &bytes.Buffer{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtText)
	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetName() == "job" {
					fmt.Println("metric ", mf.GetName(), m, "already contains a job label")
				}
			}
		}
		if err := enc.Encode(mf); err != nil {
			fmt.Println("encode error", err)
		}
	}
	return buf.String()
}

package collector

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

type Instance struct {
	R        *prometheus.Registry
	C        *NodeCollector
	job      string
	instance string
}

var instance *Instance

func GetInstance() *Instance {
	if instance == nil {
		instance = &Instance{
			R:        prometheus.NewRegistry(),
			C:        NewNodeCollector(),
			job:      "defaultJobName",
			instance: "defaultInstanceName",
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

func (ins *Instance) PushMetrics(gateway string, data string) error {
	sr := strings.NewReader(data)
	br := bufio.NewReader(sr)
	var url string
	if gateway[len(gateway)-1] == '/' {
		url = gateway + "metrics/job/" + ins.job + "/instance/" + ins.instance
	} else {
		url = gateway + "/metrics/job/" + ins.job + "/instance/" + ins.instance
	}
	req, err := http.NewRequest(http.MethodPost, url, br)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := ioutil.ReadAll(resp.Body)
		errStr := fmt.Sprintf("unexpected status code %s, PushGateway url = %s, body = %s.", resp.StatusCode, url, string(body))
		return errors.New(errStr)
	}
	return nil
}

func (ins *Instance) SetJob(job string) {
	ins.job = job
}

func (ins *Instance) GetJob() string {
	return ins.job
}

func (ins *Instance) SetInstance(instance string) {
	ins.instance = instance
}

func (ins *Instance) GetInstance() string {
	return ins.instance
}

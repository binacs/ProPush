package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/binacs/ProPush/collector"
)

var job, instance, endpoint string

func init() {
	flag.StringVar(&job, "job", "defaultJobName", "Job name")
	flag.StringVar(&instance, "instance", "defaultInstanceName", "Instance name")
	flag.StringVar(&endpoint, "endpoint", "http://127.0.0.1:9091", "Push gateway endpoints")
}

func main() {
	flag.Parse()

	ins := collector.GetInstance()
	ins.SetJob(job)
	hostname, err := os.Hostname();
	if err != nil {
		fmt.Println("os.Hostname err:", err, " use instance name")
		ins.SetInstance(instance)
	} else {
		ins.SetInstance(hostname)
	}
	
	log.Println("job=", ins.GetJob(), "instance=", ins.GetInstance(), "endpoint=", endpoint)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				err := ins.PushMetrics(endpoint, ins.GetMetrics())
				if err != nil {
					log.Println("PushMetrics err:", err)
				} else {
					log.Println("PushMetrics succeed")
				}
			}
		}
	}()
	RunForever()
}

func RunForever() {
	TrapSignal(func() {
	})
}

func TrapSignal(cb func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			fmt.Printf("captured %v, bye ^-^ \n", sig)
			if cb != nil {
				cb()
			}
			time.Sleep(2333 * time.Millisecond)
			os.Exit(1)
		}
	}()
	select {}
}

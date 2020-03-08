package main

import (
	"fmt"

	"github.com/BinacsLee/ProPush/collector"
)

func main() {
	ins := collector.GetInstance()
	fmt.Println(ins.GetMetrics())
	err := ins.PushMetrics("http://127.0.0.1:9091", ins.GetMetrics())
	if err != nil {
		fmt.Println("PushMetrics err:", err)
	}
}

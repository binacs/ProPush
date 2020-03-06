package main

import (
	"fmt"

	"github.com/BinacsLee/ProPush/collector"
)

func main() {
	ins := collector.GetInstance()
	fmt.Println(ins.GetMetrics())
}

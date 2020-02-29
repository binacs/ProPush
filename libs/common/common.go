package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunForever() {
	TrapSignal(func() {
	})
}

func TrapSignal(cb func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			fmt.Printf("captured %v, exiting...\n", sig)
			if cb != nil {
				cb()
			}
			time.Sleep(2333 * time.Millisecond)
			os.Exit(1)
		}
	}()
	select {}
}

package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var quit_chans []chan<- bool

func init() {
	quit_chans = make([]chan<- bool, 0, 2)
}

func signalHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		sig := <-c
		switch sig {
		case syscall.SIGINT, syscall.SIGQUIT:
			logger.Printf("Terminating\n")
			for _, ch := range quit_chans {
				go func(ch chan<- bool) {
					ch <- true
				}(ch)
			}
			return
		case syscall.SIGUSR1:
			logger.Printf("Reopen log file")
			reopenLogger()
		case syscall.SIGUSR2:
			buffer := make([]byte, 256*1024)
			runtime.Stack(buffer, true)
			logger.Printf(string(buffer))
		}
	}

}

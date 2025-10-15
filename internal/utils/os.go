package utils

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}

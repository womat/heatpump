package main

import (
	"os"
	"os/signal"
	"syscall"

	"heatpump/global"
	_ "heatpump/pkg/config"
	"heatpump/pkg/debug"
	_ "heatpump/pkg/webservice"
)

func main() {
	debug.SetDebug(global.Config.Debug.File, global.Config.Debug.Flag)

	// capture exit signals to ensure resources are released on exit.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	// wait for am os.Interrupt signal (CTRL C)
	sig := <-quit
	debug.InfoLog.Printf("Got %s signal. Aborting...\n", sig)
	os.Exit(1)
}

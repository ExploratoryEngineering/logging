package main

import (
	"time"

	"github.com/ExploratoryEngineering/logging"
)

func main() {
	logs := logging.NewMemoryLoggers(1024)
	logging.EnableMemoryLogger(logs)
	logging.SetLogLevel(logging.DebugLevel)

	go func() {
		c := 1
		for {
			logging.Debug("This is debug %d", c)
			if c%11 == 0 {
				logging.Info("This is dog %d", c)
			}
			if c%21 == 0 {
				logging.Warning("This is warning %d", c)
			}
			if c%51 == 0 {
				logging.Error("This is dog %d", c)
			}
			time.Sleep(35 * time.Millisecond)
			c++
		}
	}()

	term := logging.NewTerminalLogger(logs)
	term.Start()
}

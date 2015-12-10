package main

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"cm/log"
)

var (
	configPath = flag.String("config", filepath.FromSlash("/etc/cm.ini"), "Config file")
	debugHTTP  = flag.String("debug-http", "", "HTTP debug interface, to enable set like: 127.0.0.1:6060")
)

func main() {

	if !flag.Parsed() {
		flag.Parse()
	}

	logger := log.New()

	bus, err := readConfigFile(*configPath, logger)
	if err != nil {
		logger.Write("FATAL", "Config error: %s\n", err.Error())
	}

	for _, reciever := range bus.recievers {
		reciever.Start()
	}
	for _, sender := range bus.senders {
		sender.Start()
	}
	for _, task := range bus.supervisorTasks {
		task.Start()
	}

	if *debugHTTP != "" {
		go func() {
			logger.Write("DEBUG", "Listen error: %s\n", http.ListenAndServe(*debugHTTP, nil).Error())
		}()
	}

	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt)
	signal.Notify(quitSignal, syscall.SIGTERM)

	ticker := time.Tick(100 * time.Millisecond)
	for {
		select {
		case <-ticker:
			bus.Flush()
		case <-quitSignal:
			logger.Write("INFO", "Bye-bye!\n")
			os.Exit(1)
		}
	}

}

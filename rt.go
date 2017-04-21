package main

import (
	"flag"
	"github.com/aerokube/rt/api"
	"github.com/aerokube/rt/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	listen      string
	confPath    string
	logConfPath string
	dataDir     string
	timeout     time.Duration
)

func init() {
	flag.StringVar(&listen, "listen", ":8080", "Network address to accept connections")
	flag.StringVar(&confPath, "conf", "config/containers.json", "configuration file path")
	flag.StringVar(&logConfPath, "log-conf", "config/container-logs.json", "Container logging configuration file")
	flag.StringVar(&dataDir, "dataDir", "data", "directory to save results to")
	flag.DurationVar(&timeout, "timeout", 2*time.Hour, "test case timeout")
	flag.Parse()
}

func cancelOnSignal(exit chan bool) {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		close(exit)
		os.Exit(0)
	}()
}

func main() {
	conf := config.NewConfig(dataDir, timeout)
	err := conf.Load(confPath, logConfPath)
	if err != nil {
		log.Fatalf("%s: %v", os.Args[0], err)
	}
	exit := make(chan bool)
	cancelOnSignal(exit)
	go api.ConsumeLaunches(conf, exit)
	go api.ConsumeTerminates(exit)
	log.Printf("Listening on %s\n", listen)
	log.Fatal(http.ListenAndServe(listen, api.Mux(exit)))
}

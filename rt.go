package main

import (
	"flag"
	"net/http"
	"log"
	"github.com/aerokube/rt/api"
	"os"
	"syscall"
	"os/signal"
)

var (
	listen   string
)

func init() {
	flag.StringVar(&listen, "listen", ":8080", "host and port to listen to")
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
	log.Printf("Listening on %s\n", listen)
	exit := make(chan bool)
	cancelOnSignal(exit)
	go api.ConsumeLaunches(exit)
	go api.ConsumeTerminates(exit)
	log.Fatal(http.ListenAndServe(listen, api.Mux(exit)))
}
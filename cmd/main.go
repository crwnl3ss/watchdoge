package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/crwnl3ss/watchdoge/pkg/config"
	"github.com/crwnl3ss/watchdoge/pkg/scheduler"
)

var configFilePath string

func main() {
	flag.StringVar(&configFilePath, "c", "/etc/watchdoge/config.yaml", "-c = /etc/watchdoge/config.yaml")
	flag.Parse()
	fd, err := os.Open(configFilePath)
	if err != nil {
		flag.PrintDefaults()
		log.Fatalf("could not read configuration file, reason: %s", err)
	}
	cfg, err := config.LoadConfig(bufio.NewReader(fd))
	if err != nil {
		log.Printf("could not load configuration file: %s\n usage:", err)
		flag.PrintDefaults()
		os.Exit(1)
	}
	log.Println("config successfull parsed, setup checks...")
	scheduler := scheduler.NewSheduler(cfg.Periodicity, cfg.RunnableChecks)
	if err := scheduler.SetUpCheckers(); err != nil {
		log.Printf("scheduler setup error: %s", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	scheduler.Run(ctx)

	if err := scheduler.TearDownCheckers(); err != nil {
		log.Printf("scheduler teardown error: %s", err)
	}
}

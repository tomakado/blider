package main

import (
	"flag"
	"github.com/ildarkarymoff/blider/change"
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/provider"
	"github.com/ildarkarymoff/blider/schedule"
	"log"
	"os"
	"path"
)

func main() {
	homeDir, _ := os.UserHomeDir()
	defaultConfigPath := path.Join(homeDir, ".blider", "config.json")

	configPath := flag.String("config", defaultConfigPath, "path to JSON file with configuration")

	flag.Parse()

	cfg, err := config.FromFile(*configPath)
	if err != nil {
		log.Fatalf("failed to load config from %s: %v", *configPath, err)
	}

	fetcher := &provider.SimpleDesktopsProvider{}
	changer := change.NewKDEChanger(cfg)

	scheduler := schedule.NewScheduler(fetcher, changer)
	if err := scheduler.Start(cfg); err != nil {
		log.Fatalf("failed to start schedule: %v", err)
	}
}

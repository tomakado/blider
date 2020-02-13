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

	configPath := flag.String(
		"config",
		defaultConfigPath,
		"path to JSON file with configuration",
	)

	flag.Parse()

	cfg, err := config.FromFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", *configPath, err)
	}

	wpProvider := &provider.SimpleDesktopsProvider{}
	cmdBuilder, err := change.ResolveBuilder(cfg)
	if err != nil {
		log.Fatalf("Failed to resolve cmdBuilder: %v", err)
	}

	scheduler := schedule.NewScheduler(wpProvider, cmdBuilder)
	if err := scheduler.Start(cfg); err != nil {
		log.Fatalf("Scheduler error: %v", err)
	}
}

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

	cmdBuilder, err := change.ResolveBuilder(cfg)
	if err != nil {
		log.Fatalf("Failed to resolve cmdBuilder: %v", err)
	}

	providers := createProviders(cfg)

	scheduler := schedule.NewScheduler(providers, cmdBuilder)
	if err := scheduler.Start(cfg); err != nil {
		log.Fatalf("Scheduler error: %v", err)
	}
}

func createProviders(cfg *config.Config) *[]provider.IProvider {
	var providers []provider.IProvider

	for name, _ := range cfg.Providers {
		var providerToAppend provider.IProvider

		if name == config.ProviderSimpleDesktops {
			providerToAppend = &provider.SimpleDesktopsProvider{}
		} else if name == config.ProviderLocalDirectory {
			providerToAppend = &provider.LocalDirectoryProvider{}
		} else {
			log.Printf("WARN Unknown provider '%s'", name)
		}

		if providerToAppend != nil {
			providers = append(providers, providerToAppend)
		}
	}

	return &providers
}

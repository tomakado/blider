package schedule

import (
	"fmt"
	"github.com/ildarkarymoff/blider/change/cmd"
	"github.com/ildarkarymoff/blider/change/cmd/builder"
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/provider"
	"github.com/ildarkarymoff/blider/repository"
	"github.com/ildarkarymoff/blider/storage"
	"log"
	"math/rand"
	"time"
)

// Scheduler is singleton (yes -_-) object that
// controls main program loop. Every period it
// triggers changeOp().
type Scheduler struct {
	config     *config.Config
	period     *time.Ticker
	providers  *[]provider.IProvider
	builder    *builder.ICmdBuilder
	repository *repository.Repository
	storage    *storage.Storage
}

func NewScheduler(
	providers *[]provider.IProvider,
	builder *builder.ICmdBuilder,
) *Scheduler {
	return &Scheduler{
		providers: providers,
		builder:   builder,
	}
}

// Start initializes Scheduler and starts provide-change loop.
// This method should be used only once.
func (s *Scheduler) Start(config *config.Config) error {
	s.config = config

	if err := s.init(); err != nil {
		return fmt.Errorf("[init] %v", err)
	}

	for _, p := range *s.providers {
		p.Init(s.config, s.repository)
	}

	if err := s.changeOp(); err != nil {
		return fmt.Errorf("[changeOp 1st time] %v", err)
	}

	for range s.period.C {
		if err := s.changeOp(); err != nil {
			return fmt.Errorf("[changeOp] %v", err)
		}
	}

	return nil
}

func (s *Scheduler) init() error {
	log.Println("Initializing scheduler...")

	period, err := s.config.Period.ToTime()
	if err != nil {
		return err
	}

	s.period = time.NewTicker(period)

	log.Println("Opening repository...")
	rep, err := repository.Open(s.config.DBPath)
	if err != nil {
		return err
	}
	s.repository = rep

	log.Println("Opening storage...")
	st, err := storage.Open(s.config, s.repository)
	if err != nil {
		return err
	}
	s.storage = st

	log.Println("Initializing builder...")
	(*s.builder).Init(s.config)

	return nil
}

// changeOp asks provider to provider image then asks builder to
// change wallpaper.
func (s *Scheduler) changeOp() error {
	log.Println("Change desktop wallpaper operation triggered")
	providerIndex := rand.Intn(len(*s.providers))
	wallpaper := (*s.providers)[providerIndex].Provide()

	// If image obtaining failed we don't want to wait another
	// period, but should try to obtain again.
	if len(wallpaper.ImgBuffer) == 0 || !wallpaper.IsLocal() {
		return s.changeOp()
	}

	log.Println("Saving image to database...")
	id, err := s.repository.AddWallpaper(wallpaper)
	if err != nil {
		log.Printf("[Save wallpaper to database] %v", err)
	}

	wallpaper.ID = id

	if !wallpaper.IsLocal() {
		log.Println("Saving image to local repository...")
		if err := s.storage.Save(wallpaper.Filename, wallpaper.ImgBuffer); err != nil {
			return err
		}
	}

	command := (*s.builder).Build(wallpaper)
	if err := cmd.Run(command); err != nil {
		return err
	}

	log.Printf(
		"Background changed to '%s' by %s (%s)",
		wallpaper.Title,
		wallpaper.Author,
		wallpaper.OriginURL,
	)

	if s.config.LocalStorageLimit != 0 {
		if err := s.storage.CleanUp(); err != nil {
			return fmt.Errorf("[storage.CleanUp] %v", err)
		}
	}

	log.Printf("Paused for %s", s.config.Period)
	return nil
}

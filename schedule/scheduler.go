package schedule

import (
	"fmt"
	"github.com/ildarkarymoff/blider/change"
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/provider"
	"github.com/ildarkarymoff/blider/storage"
	"log"
	"time"
)

type Scheduler struct {
	config  *config.Config
	period  *time.Ticker
	fetcher *provider.IProvider
	changer *change.IChanger
	storage *storage.Storage
}

func NewScheduler(
	fetcher provider.IProvider,
	changer change.IChanger,
) *Scheduler {
	return &Scheduler{
		fetcher: &fetcher,
		changer: &changer,
	}
}

func (s *Scheduler) Start(config *config.Config) error {
	s.config = config

	if err := s.init(); err != nil {
		return fmt.Errorf("[init] %v", err)
	}

	(*s.fetcher).Init(s.config, s.storage)

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

	log.Println("Opening storage...")
	st, err := storage.Open(s.config.DBPath)
	if err != nil {
		return nil
	}

	s.storage = st

	return nil
}

func (s *Scheduler) changeOp() error {
	log.Println("Change desktop wallpaper operation triggered")
	wallpaper := (*s.fetcher).Provide()
	if err := (*s.changer).Change(wallpaper); err != nil {
		return err
	}

	log.Printf(
		"Background changed to '%s' by %s (%s)",
		wallpaper.Title,
		wallpaper.Author,
		wallpaper.OriginURL,
	)

	log.Printf("Paused for %s", s.config.Period)
	return nil
}

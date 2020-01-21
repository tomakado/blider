package scheduler

import (
	"github.com/ildarkarymoff/blider/blider"
	"github.com/ildarkarymoff/blider/blider/changer"
	"github.com/ildarkarymoff/blider/blider/fetcher"
	"github.com/ildarkarymoff/blider/blider/storage"
	"time"
)

type Scheduler struct {
	config       *blider.Config
	fetchTicker  *time.Ticker
	changeTicker *time.Ticker
	fetcher      *fetcher.IFetcher
	changer      *changer.IChanger
	storage      *storage.Storage
}

func New(
	fetcher *fetcher.IFetcher,
	changer *changer.IChanger,
) *Scheduler {
	return &Scheduler{
		fetcher: fetcher,
		changer: changer,
	}
}

func (s *Scheduler) Start(config *blider.Config) error {
	s.config = config

	if err := s.init(); err != nil {
		return err
	}

	// TODO Start goroutines here...

	return nil
}

func (s *Scheduler) init() error {
	fetchInterval, err := s.config.FetchPeriod.ToTime()
	if err != nil {
		return err
	}

	changeInterval, err := s.config.ChangePeriod.ToTime()
	if err != nil {
		return err
	}

	s.fetchTicker = time.NewTicker(fetchInterval)
	s.changeTicker = time.NewTicker(changeInterval)

	st, err := storage.Open(s.config.DBPath)
	if err != nil {
		return nil
	}

	s.storage = st

	return nil
}

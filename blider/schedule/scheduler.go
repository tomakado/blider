package schedule

import (
	"github.com/ildarkarymoff/blider/blider/change"
	config2 "github.com/ildarkarymoff/blider/blider/config"
	"github.com/ildarkarymoff/blider/blider/fetch"
	"github.com/ildarkarymoff/blider/blider/storage"
	"github.com/zenthangplus/goccm"
	"log"
	"math/rand"
	"reflect"
	"time"
)

type Scheduler struct {
	config       *config2.Config
	fetchTicker  *time.Ticker
	changeTicker *time.Ticker
	fetcher      *fetch.IFetcher
	changer      *change.IChanger
	storage      *storage.Storage
}

func NewScheduler(
	fetcher fetch.IFetcher,
	changer change.IChanger,
) *Scheduler {
	return &Scheduler{
		fetcher: &fetcher,
		changer: &changer,
	}
}

func (s *Scheduler) Start(config *config2.Config) error {
	s.config = config

	if err := s.init(); err != nil {
		return err
	}

	wallpapers := map[string]*storage.Wallpaper{}

	manager := goccm.New(2)

	go func() {
		manager.Wait()

		log.Println("Starting fetch ticker...")

		(*s.fetcher).Init(config)

		for range s.fetchTicker.C {
			fetched := (*s.fetcher).Fetch(s.config.MaxFetchPages)
			if err := s.storage.AddWallpapers(fetched); err != nil {
				log.Printf("[Add wallpapers batch] %v", err)
			}

			for _, w := range fetched {
				wallpapers[w.OriginURL] = w
			}
		}

		manager.Done()
	}()

	go func() {
		manager.Wait()

		log.Println("Starting change ticker...")

		for len(wallpapers) == 0 {
		}

		for range s.changeTicker.C {
			wallpaper := pickRandomWallpaper(&wallpapers)
			if err := (*s.changer).Change(wallpaper); err != nil {
				log.Printf("[Change wallpaper to %s] %v", wallpaper.Filename, err)
			}
		}

		manager.Done()
	}()

	manager.WaitAllDone()

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

func pickRandomWallpaper(wallpapers *map[string]*storage.Wallpaper) *storage.Wallpaper {
	keys := reflect.ValueOf(wallpapers).MapKeys()
	wallpaperAlias := keys[rand.Intn(len(keys))].String()

	return (*wallpapers)[wallpaperAlias]
}

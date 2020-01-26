package schedule

import (
	"fmt"
	"github.com/ildarkarymoff/blider/change"
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/provider"
	"github.com/ildarkarymoff/blider/storage"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

// Scheduler is singleton (yes -_-) object that
// controls main program loop. Every period it
// triggers changeOp().
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

// Start initializes Scheduler and starts provide-change loop.
// This method should be used only once.
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

// changeOp asks provider to provider image then asks changer to
// change wallpaper.
func (s *Scheduler) changeOp() error {
	log.Println("Change desktop wallpaper operation triggered")
	wallpaper := (*s.fetcher).Provide()

	log.Println("Saving image to database...")
	id, err := s.storage.AddWallpaper(wallpaper)
	if err != nil {
		log.Printf("[Save wallpaper to database] %v", err)
	}

	wallpaper.ID = id

	if wallpaper.ID == 0 {
		return nil
	}

	log.Println("Saving image to local storage...")
	s.saveImage(wallpaper.Filename, wallpaper.ImgBuffer)

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

func (s *Scheduler) saveImage(filename string, image []byte) {
	if _, err := os.Stat(s.config.LocalStoragePath); os.IsNotExist(err) {
		if err := os.MkdirAll(s.config.LocalStoragePath, os.ModePerm); err != nil {
			log.Fatalf(
				"Failed to create images directory %s: %v",
				s.config.LocalStoragePath,
				err,
			)
		}
	}

	filepath := path.Join(s.config.LocalStoragePath, filename)
	if err := ioutil.WriteFile(filepath, image, os.ModePerm); err != nil {
		log.Fatalf("Failed to write image to %s: %v", filepath, err)
	}

	log.Printf("Saved image to %s", filepath)

}

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
	"time"
)

// Scheduler is singleton (yes -_-) object that
// controls main program loop. Every period it
// triggers changeOp().
type Scheduler struct {
	config     *config.Config
	period     *time.Ticker
	provider   *provider.IProvider
	builder    *builder.ICmdBuilder
	repository *repository.Repository
	storage    *storage.Storage
}

func NewScheduler(
	provider provider.IProvider,
	builder *builder.ICmdBuilder,
) *Scheduler {
	return &Scheduler{
		provider: &provider,
		builder:  builder,
	}
}

// Start initializes Scheduler and starts provide-change loop.
// This method should be used only once.
func (s *Scheduler) Start(config *config.Config) error {
	s.config = config

	if err := s.init(); err != nil {
		return fmt.Errorf("[init] %v", err)
	}

	(*s.provider).Init(s.config, s.repository)

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
	//defer s.repository.Close()

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
	wallpaper := (*s.provider).Provide()

	// If image obtaining failed we don't want to wait another
	// period, but should try to obtain again.
	if len(wallpaper.ImgBuffer) == 0 {
		return s.changeOp()
	}

	log.Println("Saving image to database...")
	id, err := s.repository.AddWallpaper(wallpaper)
	if err != nil {
		log.Printf("[Save wallpaper to database] %v", err)
	}

	wallpaper.ID = id

	log.Println("Saving image to local repository...")
	if err := s.storage.Save(wallpaper.Filename, wallpaper.ImgBuffer); err != nil {
		return err
	}
	//s.saveImage(wallpaper.Filename, wallpaper.ImgBuffer)

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

//func (s *Scheduler) saveImage(filename string, image []byte) {
//	if _, err := os.Stat(s.config.LocalStoragePath); os.IsNotExist(err) {
//		if err := os.MkdirAll(s.config.LocalStoragePath, os.ModePerm); err != nil {
//			log.Fatalf(
//				"Failed to create images directory %s: %v",
//				s.config.LocalStoragePath,
//				err,
//			)
//		}
//	}
//
//	wpPath := path.Join(s.config.LocalStoragePath, filename)
//	if err := ioutil.WriteFile(wpPath, image, os.ModePerm); err != nil {
//		log.Fatalf("Failed to write image to %s: %v", wpPath, err)
//	}
//
//	log.Printf("Saved image to %s", wpPath)
//
//}
//
//func (s *Scheduler) cleanupLocalStorage() error {
//	log.Println("Checking local repository...")
//	files, err := ioutil.ReadDir(s.config.LocalStoragePath)
//	if err != nil {
//		return err
//	}
//
//	if len(files) > s.config.LocalStorageLimit {
//		log.Println("Local repository limit exceeded. Cleaning up...")
//		wallpapers, err := s.repository.GetWallpapers()
//		if err != nil {
//			return err
//		}
//
//		if len(wallpapers) < s.config.LocalStorageLimit {
//			return fmt.Errorf(
//				"local repository size (%d) and wallpapers count (%d) in DB mismatch",
//				len(files),
//				len(wallpapers),
//			)
//		}
//
//		for i := s.config.LocalStorageLimit + 1; i < len(wallpapers); i++ {
//			wpPath := filepath.Join(s.config.LocalStoragePath, wallpapers[i].Filename)
//
//			stat, err := os.Stat(wpPath)
//			wpIsNotExist := os.IsNotExist(err)
//			if wpIsNotExist || stat.IsDir() {
//				continue
//			}
//
//			log.Printf("Removing '%s'...", wallpapers[i].Filename)
//			if err := os.Remove(wpPath); err != nil {
//				return fmt.Errorf("[Remove '%s'] %v", wallpapers[i].Filename, err)
//			}
//		}
//	}
//
//	return err
//}

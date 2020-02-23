package storage

import (
	"fmt"
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/repository"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

// Storage is object for managing local images storage.
type Storage struct {
	config     *config.Config
	repository *repository.Repository
}

// Open checks if local storage directory exists. If not it creates directory
// corresponding to config.LocalStoragePath.
func Open(config *config.Config, repository *repository.Repository) (*Storage, error) {
	if _, err := os.Stat(config.LocalStoragePath); os.IsNotExist(err) {
		if err := os.MkdirAll(config.LocalStoragePath, os.ModePerm); err != nil {
			return &Storage{},
				fmt.Errorf(
					"failed to create images directory %s: %v",
					config.LocalStoragePath,
					err,
				)
		}
	}

	if _, err := os.Open(config.LocalStoragePath); os.IsPermission(err) {
		return &Storage{}, fmt.Errorf(
			"failed to open images directory %s: %v",
			config.LocalStoragePath, err,
		)
	}

	return &Storage{
		config:     config,
		repository: repository,
	}, nil
}

// Save tries to write images bytes to specified file.
// Returns error on failure.
func (s *Storage) Save(filename string, image []byte) error {
	wpPath := path.Join(s.config.LocalStoragePath, filename)
	if err := ioutil.WriteFile(wpPath, image, os.ModePerm); err != nil {
		return fmt.Errorf("failed to write image to %s: %v", wpPath, err)
	}

	log.Printf("Saved image to %s", wpPath)
	return nil
}

// CleanUp makes locally saved images amount to be limited
// to LocalStorageLimit parameter in configuration.
// CleanUp selects images for deleting from disk based on
// information from SQLite database.
func (s *Storage) CleanUp() error {
	log.Println("Checking local repository...")
	files, err := ioutil.ReadDir(s.config.LocalStoragePath)
	if err != nil {
		return err
	}

	if len(files) <= s.config.LocalStorageLimit {
		return nil
	}

	log.Println("Local repository limit exceeded. Cleaning up...")

	for _, wp := range files {
		wpFilename := wp.Name()
		wpPath := filepath.Join(s.config.LocalStoragePath, wpFilename)

		stat, err := os.Stat(wpPath)
		wpIsNotExist := os.IsNotExist(err)
		if wpIsNotExist || stat.IsDir() {
			continue
		}

		log.Printf("Removing '%s'...", wpFilename)
		if err := os.Remove(wpPath); err != nil {
			return fmt.Errorf("[Remove '%s'] %v", wpFilename, err)
		}
	}

	return nil
}

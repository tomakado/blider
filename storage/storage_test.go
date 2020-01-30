package storage

import (
	"fmt"
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/repository"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"path/filepath"
	"testing"
)

var (
	cfg *config.Config
	rep *repository.Repository
)

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	dbPath := filepath.Join(wd, "blider_test.sqlite")
	localStoragePath := path.Join(wd, ".blider_test")

	cfg = &config.Config{
		DBPath:           dbPath,
		LocalStoragePath: localStoragePath,
	}

	rep, err = repository.Open(dbPath)
	if err != nil {
		fmt.Printf("Failed to open repository: %v\n", err)
		os.Exit(1)
	}

	m.Run()

	_ = os.Remove(dbPath)
	_ = os.Remove(localStoragePath)
}

func TestOpen(t *testing.T) {
	storage, err := Open(cfg, rep)
	assert.NoError(t, err)
	assert.NotEmpty(t, storage)

	wrongCfg := config.NewDefault()
	wrongCfg.LocalStoragePath = "/root/.blider"

	storage, err = Open(wrongCfg, rep)
	assert.Error(t, err)
	assert.Empty(t, storage)
}

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
	"time"
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
		DBPath:            dbPath,
		LocalStoragePath:  localStoragePath,
		LocalStorageLimit: 5,
	}

	rep, err = repository.Open(dbPath)
	if err != nil {
		fmt.Printf("Failed to open repository: %v\n", err)
		os.Exit(1)
	}

	m.Run()

	_ = os.Remove(dbPath)
	_ = os.RemoveAll(localStoragePath)
}

func TestOpen(t *testing.T) {
	storage, err := Open(cfg, rep)
	assert.NotEmpty(t, storage)
	assert.NoError(t, err)

	wrongCfg := config.NewDefault()
	wrongCfg.LocalStoragePath = "/root/.blider"

	storage, err = Open(wrongCfg, rep)
	assert.Empty(t, storage)
	assert.Error(t, err)
}

func TestStorage_Save(t *testing.T) {
	const testFilename = "test.png"

	storage, err := Open(cfg, rep)
	assert.NotEmpty(t, storage)
	assert.NoError(t, err)

	assert.NoError(t, storage.Save(testFilename, []byte{}))

	// Negative test case is not provided here
	// because test for Storage.Open() covers it.
}

func TestStorage_CleanUp(t *testing.T) {
	const testFilenameFmt = "test_%d.png"

	storage, err := Open(cfg, rep)
	assert.NotEmpty(t, storage)
	assert.NoError(t, err)

	filename := fmt.Sprintf(testFilenameFmt, -1)

	wallpaper := &repository.Wallpaper{
		Filename:       filename,
		FetchTimestamp: uint(time.Now().Unix()),
		Title:          "Test",
		Author:         "go test",
		AuthorURL:      "https://golang.org",
		ImgBuffer:      []byte{},
	}

	_, err = rep.AddWallpaper(wallpaper)
	assert.NoError(t, err)

	assert.NoError(t, storage.Save(filename, []byte{}))

	assert.NoError(t, storage.CleanUp())

	for i := 0; i < 10; i++ {
		filename = fmt.Sprintf(testFilenameFmt, i)
		assert.NoError(t, storage.Save(filename, []byte{}))
	}

	assert.Error(t, storage.CleanUp())
}

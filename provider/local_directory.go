package provider

import (
	"fmt"
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/repository"
	"io/ioutil"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type LocalDirectoryProvider struct {
	config     *config.Config
	repository *repository.Repository
	path       string
}

func (p *LocalDirectoryProvider) Init(cfg *config.Config, repository *repository.Repository) {
	p.config = cfg
	p.repository = repository
	p.path = cfg.LocalStoragePath

	providerConfig, cfgOk := cfg.Providers[config.ProviderLocalDirectory]
	if cfgOk {
		path, pathOk := (*providerConfig)["path"].(string)
		if pathOk {
			p.path = path
		}
	}
}

func (p *LocalDirectoryProvider) Provide() *repository.Wallpaper {
	files, err := ioutil.ReadDir(p.path)
	if err != nil {
		fmt.Printf("ERROR [ReadDir(%s)] %v", p.path, err)
		return nil
	}

	images := selectImagesOnly(files)
	if len(images) == 0 {
		fmt.Printf(
			"ERROR [selectImagesOnly] No images found in '%s'",
			p.path,
		)

		return nil
	}

	imgIndex := rand.Intn(len(images))
	selectedImg := images[imgIndex]
	imgFilename := selectedImg.Name()
	title := imgFilename[0 : len(imgFilename)-len(filepath.Ext(imgFilename))]

	userName := "Someone"
	user, err := user.Current()
	if err != nil {
		fmt.Printf("WARN [user.Current] %v", err)
	} else {
		userName = user.Name
	}

	return &repository.Wallpaper{
		OriginURL:      fmt.Sprintf("file://%s", selectedImg.Name()),
		Filename:       selectedImg.Name(),
		FetchTimestamp: uint(time.Now().Unix()),
		Title:          title,
		Author:         userName,
		AuthorURL:      "http://localhost",
	}
}

func selectImagesOnly(files []os.FileInfo) []os.FileInfo {
	var selected []os.FileInfo

	for _, f := range files {
		if isImage((f).Name()) {
			selected = append(selected, f)
		}
	}

	return selected
}

type extList []string

func (l *extList) contains(ext string) bool {
	for _, e := range *l {
		if e == ext {
			return true
		}
	}

	return false
}

var (
	allowedExtensions = extList{"png", "jpg", "jpeg", "bmp"}
)

func isImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return allowedExtensions.contains(ext)
}

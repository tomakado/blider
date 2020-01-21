package fetcher

import (
	"github.com/ildarkarymoff/kde-simpledesktops/blider"
	"github.com/ildarkarymoff/kde-simpledesktops/blider/storage"
)

type IFetcher interface {
	Init(config *blider.Config)
	Fetch(limit int) []*storage.Wallpaper
}

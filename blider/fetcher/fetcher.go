package fetcher

import (
	"github.com/ildarkarymoff/blider/blider"
	"github.com/ildarkarymoff/blider/blider/storage"
)

type IFetcher interface {
	Init(config *blider.Config)
	Fetch(limit int) []*storage.Wallpaper
}

package fetch

import (
	"github.com/ildarkarymoff/blider/blider/config"
	"github.com/ildarkarymoff/blider/blider/storage"
)

type IFetcher interface {
	Init(config *config.Config)
	Fetch(int) []*storage.Wallpaper
}

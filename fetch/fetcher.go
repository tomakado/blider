package fetch

import (
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/storage"
)

type IFetcher interface {
	Init(config *config.Config, storage *storage.Storage)
	Fetch() *storage.Wallpaper
}

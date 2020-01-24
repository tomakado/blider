package provider

import (
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/storage"
)

type IProvider interface {
	Init(config *config.Config, storage *storage.Storage)
	Provide() *storage.Wallpaper
}

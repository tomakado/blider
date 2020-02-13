package provider

import (
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/repository"
)

// IProvider is an interface for creating different image providers.
// Generally speaking, IProvider can generate images on the fly,
// not only download them from external sources,.
type IProvider interface {
	// Init opens SQLite database connection and does
	// some provider-specific stuff.
	Init(config *config.Config, storage *repository.Repository)

	// Provide is a main method for each provider that
	// must obtain or generate image. Also at current
	// code state it save wallpaper in database, but
	// saving to database logic will be moved to separate
	// module in future versions.
	Provide() *repository.Wallpaper
}

package provider

import (
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/repository"
)

type LocalDirectoryProvider struct {
	config     *config.Config
	repository *repository.Repository
}

func (p *LocalDirectoryProvider) Init(config *config.Config, repository *repository.Repository) {
	p.config = config
	p.repository = repository
}

//func (p *LocalDirectoryProvider) Provide() *repository.Wallpaper {
//
//}
//
//func (p )

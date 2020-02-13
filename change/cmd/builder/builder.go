package builder

import (
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/repository"
	"os/exec"
)

type ICmdBuilder interface {
	Init(config *config.Config)
	Build(wallpaper *repository.Wallpaper) *exec.Cmd
}

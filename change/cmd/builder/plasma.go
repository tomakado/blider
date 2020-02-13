package builder

import (
	"fmt"
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/repository"
	"os/exec"
	"path/filepath"
)

const (
	scriptFmt = `var allDesktops = desktops();
		print (allDesktops);
		for (i=0;i<allDesktops.length;i++) {{
			d = allDesktops[i];
			d.wallpaperPlugin = "org.kde.image";
			d.currentConfigGroup = Array("Wallpaper",
										 "org.kde.image",
										 "General");
			d.writeConfig("Image", "file://%s")
		}}`
)

type PlasmaCmdBuilder struct {
	config *config.Config
}

func (b *PlasmaCmdBuilder) Init(config *config.Config) {
	b.config = config
}

func (b *PlasmaCmdBuilder) Build(wallpaper *repository.Wallpaper) *exec.Cmd {
	imgPath := filepath.Join(b.config.LocalStoragePath, wallpaper.Filename)
	script := fmt.Sprintf(scriptFmt, imgPath)
	return exec.Command(
		"qdbus",
		"org.kde.plasmashell",
		"/PlasmaShell",
		"evaluateScript",
		script,
	)
}

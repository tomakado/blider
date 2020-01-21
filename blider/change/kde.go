package change

import (
	"fmt"
	config2 "github.com/ildarkarymoff/blider/blider/config"
	"github.com/ildarkarymoff/blider/blider/storage"
	"log"
	"os/exec"
	"path"
)

const (
	changeBgCmdFormat = `dbus-send --session --dest=org.kde.plasmashell --type=method_call /PlasmaShell org.kde.PlasmaShell.evaluateScript 'string:
		var Desktops = desktops();
		for (var i = 0; i< Desktops.length ; i++) {
			d = Desktops[i];
			d.wallpaperPlugin = "org.kde.image";
			d.currentConfigGroup = Array("Wallpaper",
			"org.kde.image",
			"General");
			d.writeConfig("Image", "file:///%s");
		}'`
)

type KDEChanger struct {
	config *config2.Config
}

func NewKDEChanger(config *config2.Config) *KDEChanger {
	return &KDEChanger{
		config: config,
	}
}

func (c KDEChanger) Change(wallpaper *storage.Wallpaper) error {
	filepath := path.Join(c.config.LocalStoragePath, wallpaper.Filename)
	command := fmt.Sprintf(changeBgCmdFormat, filepath)
	cmd := exec.Command(command)

	if err := cmd.Run(); err != nil {
		return err
	}

	log.Printf("Background changed to %s", filepath)
	return nil
}

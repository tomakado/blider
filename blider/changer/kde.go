package changer

import (
	"database/sql"
	"fmt"
	"github.com/ildarkarymoff/kde-simpledesktops/blider"
	"github.com/ildarkarymoff/kde-simpledesktops/blider/storage"
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
	db     *sql.DB
	config *blider.Config
}

func NewKDEChanger(db *sql.DB, config *blider.Config) *KDEChanger {
	return &KDEChanger{
		db:     db,
		config: config,
	}
}

func (c KDEChanger) Change(wallpaper *storage.Wallpaper) error {
	filepath := path.Join(c.config.StoragePath, wallpaper.Filename)
	command := fmt.Sprintf(changeBgCmdFormat, filepath)
	cmd := exec.Command(command)

	if err := cmd.Run(); err != nil {
		return err
	}

	log.Printf("Background changed to %s", filepath)
	return nil
}

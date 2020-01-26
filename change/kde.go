package change

import (
	"bytes"
	"fmt"
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/storage"
	"log"
	"os/exec"
	"path"
	"strings"
	"syscall"
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

// KDEChanger is IChanger implementation for KDE Plasma Desktop Environment.
type KDEChanger struct {
	config *config.Config
}

func NewKDEChanger(config *config.Config) *KDEChanger {
	return &KDEChanger{
		config: config,
	}
}

// Change calls special Plasma script to change desktop wallpaper providing
// path to image using information from config and storage.Wallpaper instance.
func (c KDEChanger) Change(wallpaper *storage.Wallpaper) error {
	filepath := path.Join(c.config.LocalStoragePath, wallpaper.Filename)
	script := fmt.Sprintf(scriptFmt, filepath)
	cmd := exec.Command(
		"qdbus",
		"org.kde.plasmashell",
		"/PlasmaShell",
		"evaluateScript",
		script,
	)

	var output bytes.Buffer
	cmd.Stdout = &output

	var errs bytes.Buffer
	cmd.Stderr = &errs

	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd.Start() error: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				log.Printf("Exit status: %d", status.ExitStatus())
			}
		} else {
			log.Fatalf("cmd.Wait() error: %v", err)
		}
	}

	outputStr := strings.TrimSpace(output.String())
	if len(outputStr) > 0 {
		log.Println(output.String())
	}

	errsStr := strings.TrimSpace(errs.String())
	if len(errsStr) > 0 {
		log.Println(errs.String())
	}

	return nil
}

package change

import (
	"bytes"
	"fmt"
	config2 "github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/storage"
	"log"
	"os/exec"
	"path"
	"strings"
	"syscall"
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
	script := fmt.Sprintf(`var allDesktops = desktops();
		print (allDesktops);
		for (i=0;i<allDesktops.length;i++) {{
			d = allDesktops[i];
			d.wallpaperPlugin = "org.kde.image";
			d.currentConfigGroup = Array("Wallpaper",
										 "org.kde.image",
										 "General");
			d.writeConfig("Image", "file://%s")
		}}
	`, filepath,
	)
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

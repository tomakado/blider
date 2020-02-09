package change

import (
	"bytes"
	"fmt"
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/repository"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type GnomeChanger struct {
	config *config.Config
}

func NewGnomeChanger(config *config.Config) *GnomeChanger {
	return &GnomeChanger{
		config: config,
	}
}

func (c GnomeChanger) Change(wallpaper *repository.Wallpaper) error {
	imgPath := filepath.Join(c.config.LocalStoragePath, wallpaper.Filename)
	// gsettings set org.gnome.desktop.background picture-uri file:///$PATH_TO_FILE
	cmd := exec.Command(
		"gsettings",
		"set",
		"org.gnome.desktop.background",
		"picture-uri",
		fmt.Sprintf("file:///%s", imgPath),
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

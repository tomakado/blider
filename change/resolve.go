package change

import (
	"errors"
	"fmt"
	"github.com/ildarkarymoff/blider/change/cmd/builder"
	"github.com/ildarkarymoff/blider/config"
	"log"
	"os"
	"runtime"
)

const (
	osLinux = "linux"

	deKde   = "KDE"
	deGnome = "GNOME"
)

var (
	supportedOS = envList{
		osLinux,
	}

	supportedDe = envList{
		deKde,
		deGnome,
	}
)

type envList []string

func (l *envList) contains(env string) bool {
	for _, e := range *l {
		if e == env {
			return true
		}
	}

	return false
}

func ResolveBuilder(config *config.Config) (*builder.ICmdBuilder, error) {
	goos := runtime.GOOS

	if !supportedOS.contains(goos) {
		return nil, fmt.Errorf(
			"OS '%s' is not supported",
			goos,
		)
	}

	if goos == "linux" {
		return resolveDesktopEnvironment(config), nil
	}

	return nil, errors.New("environment is not supported")
}

func resolveDesktopEnvironment(config *config.Config) *builder.ICmdBuilder {
	builders := map[string]builder.ICmdBuilder{
		deKde:   &builder.PlasmaCmdBuilder{},
		deGnome: &builder.GnomeCmdBuilder{},
	}

	de := os.Getenv("XDG_CURRENT_DESKTOP")
	cmdBuilder, ok := builders[de]

	if !ok {
		log.Println(
			"Failed to detect desktop environment. Switching to Gnome...",
		)
		cmdBuilder = builders[deGnome]
	}

	log.Printf("Detected desktop environment: %s", de)

	return &cmdBuilder
}

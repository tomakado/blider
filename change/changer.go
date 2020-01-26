package change

import "github.com/ildarkarymoff/blider/storage"

// IChanger is an interface for creating desktop
// wallpaper changers for different desktop
// environments and operating systems.
type IChanger interface {
	// Change is a main method of each changer.
	// Returns error only on failure.
	Change(wallpaper *storage.Wallpaper) error
}

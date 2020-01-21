package changer

import "github.com/ildarkarymoff/kde-simpledesktops/blider/storage"

type IChanger interface {
	Change(wallpaper *storage.Wallpaper) error
}

package change

import "github.com/ildarkarymoff/blider/storage"

type IChanger interface {
	Change(wallpaper *storage.Wallpaper) error
}

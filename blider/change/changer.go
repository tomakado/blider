package change

import "github.com/ildarkarymoff/blider/blider/storage"

type IChanger interface {
	Change(wallpaper *storage.Wallpaper) error
}

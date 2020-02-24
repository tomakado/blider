package repository

import "strings"

type Query struct {
	Filters    *Filters
	Pagination *Pagination
}

type Filters struct {
	OriginURLEquals   string
	OriginURLContains string

	OriginTypes OriginTypeList

	FilenameEquals   string
	FilenameContains string

	FetchTimestampEquals      uint
	FetchTimestampGreaterThan uint
	FetchTimestampLessThen    uint

	TitleEquals   string
	TitleContains string

	AuthorEquals   string
	AuthorContains string

	AuthorURLEquals   string
	AuthorURLContains string
}

type OriginType string

func DetectOriginType(wallpaper *Wallpaper) OriginType {
	if wallpaper.IsLocal() {
		return OriginTypeLocal
	}

	return OriginTypeRemote
}

type OriginTypeList []OriginType

func (l *OriginTypeList) Contains(t OriginType) bool {
	for _, type_ := range *l {
		if type_ == t {
			return true
		}
	}

	return false
}

const (
	OriginTypeLocal  = "local"
	OriginTypeRemote = "remote"
)

func (f *Filters) Fits(wallpaper *Wallpaper) bool {
	fits := true

	if f.OriginURLEquals != "" {
		fits = fits && wallpaper.OriginURL == f.OriginURLEquals
	} else if f.OriginURLContains != "" {
		fits = fits && strings.Contains(
			wallpaper.OriginURL,
			f.OriginURLContains,
		)
	}

	if len(f.OriginTypes) != 0 {
		fits = fits && f.OriginTypes.Contains(
			DetectOriginType(wallpaper),
		)
	}

	if f.FilenameEquals != "" {
		fits = fits && wallpaper.Filename == f.FilenameEquals
	} else if f.FilenameContains != "" {
		fits = fits && strings.Contains(
			wallpaper.Filename,
			f.FilenameContains,
		)
	}

	if f.FetchTimestampEquals != 0 {
		fits = fits &&
			wallpaper.FetchTimestamp == f.FetchTimestampEquals
	} else if f.FetchTimestampGreaterThan != 0 {
		fits = fits &&
			wallpaper.FetchTimestamp > f.FetchTimestampGreaterThan
	} else if f.FetchTimestampLessThen != 0 {
		fits = fits &&
			wallpaper.FetchTimestamp < f.FetchTimestampLessThen
	}

	if f.TitleEquals != "" {
		fits = fits && wallpaper.Title == f.TitleEquals
	} else if f.TitleContains != "" {
		fits = fits && strings.Contains(
			wallpaper.Title,
			f.TitleContains,
		)
	}

	if f.AuthorEquals != "" {
		fits = fits && wallpaper.Author == f.AuthorEquals
	} else if f.AuthorContains != "" {
		fits = fits && strings.Contains(
			wallpaper.Author,
			f.AuthorContains,
		)
	}

	if f.AuthorURLEquals != "" {
		fits = fits && wallpaper.AuthorURL == f.AuthorURLEquals
	} else if f.AuthorURLContains != "" {
		fits = fits && strings.Contains(
			wallpaper.AuthorURL,
			f.AuthorURLContains,
		)
	}

	return fits
}

type Pagination struct {
	Offset uint
	Limit  uint
}

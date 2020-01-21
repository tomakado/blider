package storage

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type Wallpaper struct {
	ID             int
	OriginURL      string
	Filename       string
	FetchTimestamp uint
}

type Storage struct {
	db               *sql.DB
	HasAnyWallpapers bool
}

func Open(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) AddWallpaper(wallpaper *Wallpaper) error {
	queryFormat := "insert into wallpapers (origin_url, filename, fetch_timestamp) values ('%s', '%s', %d)"
	query := fmt.Sprintf(
		queryFormat,
		wallpaper.OriginURL,
		wallpaper.Filename,
		wallpaper.FetchTimestamp,
	)
	if _, err := s.db.Exec(query); err != nil {
		return err
	}

	s.HasAnyWallpapers = true

	return nil
}

func (s *Storage) AddWallpapers(wallpapers []*Wallpaper) error {
	for _, w := range wallpapers {
		presented, err := s.IsOriginURLAlreadyPresented(w.OriginURL)
		if err != nil {
			log.Printf("[Check if is wallpaper already downloaded earlier] %v", err)
			continue
		}

		if !presented {
			if err := s.AddWallpaper(w); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Storage) GetWallpaper(id int) (*Wallpaper, error) {
	queryFormat := "select * from wallpapers where id = %d"
	query := fmt.Sprintf(queryFormat, id)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallpapers []*Wallpaper

	for rows.Next() {
		w := &Wallpaper{}
		if err := rows.Scan(&w.ID,
			&w.OriginURL,
			&w.Filename,
			&w.FetchTimestamp,
		); err != nil {
			return nil, err
		}
		wallpapers = append(wallpapers, w)
	}

	if len(wallpapers) == 0 {
		return nil, errors.New("wallpaper not found")
	}

	return wallpapers[0], nil
}

func (s *Storage) ClearStorage() error {
	//noinspection SqlWithoutWhere
	_, err := s.db.Exec("delete from wallpapers")
	return err
}

func (s *Storage) IsOriginURLAlreadyPresented(originUrl string) (bool, error) {
	queryFormat := "select * from wallpapers where origin_url = \"%s\""
	query := fmt.Sprintf(queryFormat, originUrl)

	rows, err := s.db.Query(query)
	if err != nil {
		return false, err
	}

	var wallpapers []*Wallpaper

	for rows.Next() {
		w := &Wallpaper{}
		if err := rows.Scan(&w.ID,
			&w.OriginURL,
			&w.Filename,
			&w.FetchTimestamp,
		); err != nil {
			return false, err
		}
		wallpapers = append(wallpapers, w)
	}

	return len(wallpapers) != 0, nil
}

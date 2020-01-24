package storage

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

type Wallpaper struct {
	ID             int64
	OriginURL      string
	Filename       string
	FetchTimestamp uint
	Title          string
	Author         string
	AuthorURL      string
}

type Storage struct {
	db *sql.DB
}

func Open(dbPath string) (*Storage, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Println("Instantiating new database...")
		if err := createDatabase(dbPath); err != nil {
			return nil, err
		}
		log.Println("Database created")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
}

func createDatabase(dbPath string) error {
	file, err := os.OpenFile(dbPath, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	_ = file.Close()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `CREATE TABLE history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		origin_url TEXT,
		filename TEXT,
		fetch_timestamp INTEGER,
		title TEXT,
		author TEXT,
		author_url TEXT
	)`
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("[Instantiate SQLite Database] %v", err)
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) AddWallpaper(wallpaper *Wallpaper) (int64, error) {
	queryFormat := `insert into history (
						origin_url,
						filename,
						fetch_timestamp,
						title,
						author,
						author_url)
					values ('%s', '%s', %d, '%s', '%s', '%s')`
	query := fmt.Sprintf(
		queryFormat,
		wallpaper.OriginURL,
		wallpaper.Filename,
		wallpaper.FetchTimestamp,
		wallpaper.Title,
		wallpaper.Author,
		wallpaper.AuthorURL,
	)

	result, err := s.db.Exec(query)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Storage) GetWallpaper(id int) (*Wallpaper, error) {
	queryFormat := "select * from history where id = %d"
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
	_, err := s.db.Exec("delete from history")
	return err
}

func (s *Storage) IsOriginURLAlreadyPresented(originUrl string) (bool, error) {
	queryFormat := "select * from history where origin_url = \"%s\""
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

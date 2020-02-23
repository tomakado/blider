package repository

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"sort"
	"strings"
)

const (
	WallapperLocal = "local"
)

type Wallpaper struct {
	ID int64
	// OriginURL is URL of web page where wallpaper has been taken.
	OriginURL string
	// Filename is name of the file in local image repository.
	Filename string
	// FetchTimestamp is a time when wallpaper has been fetched and downloaded.
	FetchTimestamp uint
	// Title is original title of wallpaper on source website.
	Title string
	// Author is full name or nickname of image publisher (optional).
	Author string
	// AuthorURL is author's homepage address (optional).
	AuthorURL string
	// ImgBuffer contains image bytes taken from provider.
	ImgBuffer []byte
}

func (w *Wallpaper) IsLocal() bool {
	return strings.HasPrefix(w.OriginURL, "file://")
}

// Repository allows other program modules to make operations with local SQLite database.
// Now it's used for storing history only.
type Repository struct {
	db *sql.DB
}

// Open tries to open SQLite connection. Returns Repository instance on success
// or error on failure.
func Open(dbPath string) (*Repository, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Println("Instantiating new database...")
		if err := createDatabase(dbPath); err != nil {
			return nil, err
		}
		log.Println("Database created")
	}

	_, err := os.OpenFile(dbPath, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &Repository{
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

// Close ...
func (r *Repository) Close() error {
	return r.db.Close()
}

// AddWallpaper ...
func (r *Repository) AddWallpaper(wallpaper *Wallpaper) (int64, error) {
	queryFormat := `insert into history (
						origin_url,
						filename,
						fetch_timestamp,
						title,
						author,
						author_url)
					values ("%s", "%s", %d, %q, %q, %q)`
	query := fmt.Sprintf(
		queryFormat,
		wallpaper.OriginURL,
		wallpaper.Filename,
		wallpaper.FetchTimestamp,
		wallpaper.Title,
		wallpaper.Author,
		wallpaper.AuthorURL,
	)

	result, err := r.db.Exec(query)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetWallpaper ...
func (r *Repository) GetWallpaper(id int) (*Wallpaper, error) {
	queryFormat := "select * from history where id = %d"
	query := fmt.Sprintf(queryFormat, id)

	rows, err := r.db.Query(query)
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

// ClearHistory ...
func (r *Repository) ClearHistory() error {
	//noinspection SqlWithoutWhere
	_, err := r.db.Exec("delete from history")
	return err
}

// IsOriginURLAlreadyPresented is legacy method used in past for checking if
// wallpaper has already downloaded earlier. Now I consider removing this.
func (r *Repository) IsOriginURLAlreadyPresented(originUrl string) (bool, error) {
	queryFormat := "select * from history where origin_url = %q"
	query := fmt.Sprintf(queryFormat, originUrl)

	rows, err := r.db.Query(query)
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
			&w.Title,
			&w.Author,
			&w.AuthorURL,
		); err != nil {
			return false, err
		}
		wallpapers = append(wallpapers, w)
	}

	return len(wallpapers) != 0, nil
}

// GetWallpapers ...
func (r *Repository) GetWallpapers() ([]*Wallpaper, error) {
	query := "select * from history"

	rows, err := r.db.Query(query)
	if err != nil {
		return []*Wallpaper{}, err
	}

	var wallpapers Wallpapers

	for rows.Next() {
		w := &Wallpaper{}
		if err := rows.Scan(&w.ID,
			&w.OriginURL,
			&w.Filename,
			&w.FetchTimestamp,
			&w.Title,
			&w.Author,
			&w.AuthorURL,
		); err != nil {
			return []*Wallpaper{}, err
		}

		wallpapers = append(wallpapers, w)
	}

	sort.Sort(sort.Reverse(wallpapers))
	return wallpapers, nil
}

type Wallpapers []*Wallpaper

func (w Wallpapers) Len() int {
	return len(w)
}

func (w Wallpapers) Less(i, j int) bool {
	return w[i].FetchTimestamp < w[j].FetchTimestamp
}

func (w Wallpapers) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

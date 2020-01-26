package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Period is a time interval between fetch & change iterations.
	Period Period `json:"period,omitempty"`
	// LocalStoragePath is path to local directory used as images storage.
	LocalStoragePath string `json:"local_storage_path"`
	// DBPath is path to SQLite databse.
	DBPath string `json:"db_path"`
	// MaxFetchPages is maximum number of pages to look at.
	// This parameter is being passed to provider and
	// can be changed in runtime. For example, SimpleDesktopsProvider
	// changes it on each iteration to optimize next
	// wallpaper search.
	MaxFetchPages int `json:"max_fetch_pages"`
}

// FromFile tries to load configuration from JSON file.
// If some of configuration fields have wrong or empty values
// FromFile sets default values.
// Returns Config instance on success and error on failure.
func FromFile(filename string) (*Config, error) {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var c *Config
	if err := json.Unmarshal(f, &c); err != nil {
		return nil, err
	}

	homeDir, _ := os.UserHomeDir()

	c.LocalStoragePath = strings.TrimSpace(c.LocalStoragePath)
	if len(c.LocalStoragePath) == 0 {
		c.LocalStoragePath = path.Join(homeDir, ".blider", "images")
	}

	c.DBPath = strings.TrimSpace(c.DBPath)
	if len(c.DBPath) == 0 {
		c.DBPath = path.Join(homeDir, ".blider", "blider.sqlite")
	}

	if c.MaxFetchPages <= 0 {
		c.MaxFetchPages = 10
	}

	return c, nil
}

// Period is a string in format "<integers>(s|m|h)"
// representing time duration. For example: 12h, 30s
// or 5h. Integer in period must be positive.
type Period string

// ToTime transforms Period to time.Duration.
// Returns time.Duration on success and error on failure.
func (p *Period) ToTime() (time.Duration, error) {
	pStr := strings.TrimSpace(string(*p))
	numVal, err := strconv.Atoi(pStr[:len(pStr)-1])
	if err != nil {
		return 0 * time.Second, err
	}

	// numVal must be positive value
	if numVal <= 0 {
		numVal = 1
	}

	var scale time.Duration
	scaleKey := pStr[len(pStr)-1:]

	switch scaleKey {
	case "h":
		scale = time.Hour
	case "m":
		scale = time.Minute
	case "s":
		scale = time.Second
	default:
		scale = time.Minute
	}

	return time.Duration(numVal) * scale, nil
}

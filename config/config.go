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
	// ChangePeriod is time period of changing wallpaper. Examples: 1h, 3m, 15s. Default: 1m (1 minute).
	Period Period `json:"period,omitempty"`
	// Mode represents way of selecting wallpaper to set: latest or random. Default: random.
	LocalStoragePath string `json:"local_storage_path"`
	DBPath           string `json:"db_path"`
	MaxFetchPages    int    `json:"max_fetch_pages"`
}

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

type Period string

// ToTime transforms string in format "<numbers>(s|m|h)" to Duration instance
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

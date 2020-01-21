package blider

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
	ChangePeriod Period `json:"change_period,omitempty"`
	// FetchPeriod is time period of fetching new wallpapers. Requires the same format like ChangePeriod.  Default: 1m.
	FetchPeriod Period `json:"fetch_period,omitempty"`
	// Mode represents way of selecting wallpaper to set: latest or random. Default: random.
	Mode               Mode   `json:"mode,omitempty"`
	StoragePath        string `json:"storage_path"`
	MaxFetchGoroutines int    `json:"max_fetch_goroutines"`
	MaxFetchPages      int    `json:"max_fetch_pages"`
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

	if !(c.Mode == ModeLatest || c.Mode == ModeRandom) {
		c.Mode = ModeRandom
	}

	c.StoragePath = strings.TrimSpace(c.StoragePath)
	if len(c.StoragePath) == 0 {
		homeDir, _ := os.UserHomeDir()
		c.StoragePath = path.Join(homeDir, ".blider", "images")
	}

	if c.MaxFetchGoroutines <= 0 {
		c.MaxFetchGoroutines = 4
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
	scaleKey := pStr[len(pStr)-2:]

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

type Mode string

const (
	ModeLatest Mode = "latest"
	ModeRandom Mode = "random"
)

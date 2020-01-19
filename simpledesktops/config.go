package simpledesktops

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// ChangePeriod is time period of changing wallpaper. Examples: 1h, 3m, 15s.
	ChangePeriod Period `json:"change_period"`
	// FetchPeriod is time period of fetching new wallpapers. Requires the same format like ChangePeriod.
	FetchPeriod Period `json:"fetch_period"`
	// Mode represents way of selecting wallpaper to set: latest or random.
	Mode Mode `json:"mode"`
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
		scale = time.Second
	}

	return time.Duration(numVal) * scale, nil
}

type Mode string

const (
	ModeLatest Mode = "latest"
	ModeRandom Mode = "random"
)

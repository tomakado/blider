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

const (
	ProviderSimpleDesktops = "simpledesktops"
	ProviderLocalDirectory = "local_directory"
)

type Config struct {
	// Period is a time interval between fetch & change iterations.
	Period Period `json:"period,omitempty"`
	// LocalStoragePath is path to local directory used as images repository.
	LocalStoragePath string `json:"local_storage_path"`
	// LocalStorageLimit is maximum amount of locally stored images.
	LocalStorageLimit int `json:"local_storage_limit"`
	// DBPath is path to SQLite databse.
	DBPath    string `json:"db_path"`
	Providers map[string]*map[string]interface{}
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

	c.Fill()

	return c, nil
}

// NewDefault ...
func NewDefault() *Config {
	c := &Config{}
	c.Fill()
	return c
}

func (c *Config) Fill() {
	homeDir, _ := os.UserHomeDir()

	c.LocalStoragePath = strings.TrimSpace(c.LocalStoragePath)
	if len(c.LocalStoragePath) == 0 {
		c.LocalStoragePath = path.Join(homeDir, ".blider", "images")
	}

	c.DBPath = strings.TrimSpace(c.DBPath)
	if len(c.DBPath) == 0 {
		c.DBPath = path.Join(homeDir, ".blider", "blider.sqlite")
	}

	if c.LocalStorageLimit < 0 {
		c.LocalStorageLimit = 100
	}
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

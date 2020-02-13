package repository

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

var (
	dbPath string
)

const (
	wrongDbPath = "/root/.blider/blider_test.sqlite"
)

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	dbPath = filepath.Join(wd, "blider_test.sqlite")

	m.Run()

	_ = os.Remove(dbPath)
}

func TestOpen(t *testing.T) {
	rep, err := Open(dbPath)
	assert.NotEmpty(t, rep)
	assert.NoError(t, err)

	rep, err = Open(wrongDbPath)
	assert.Empty(t, rep)
	assert.Error(t, err)
}

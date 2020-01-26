package provider

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/storage"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	entryPointURL     = "http://simpledesktops.com/browse/%d"
	simpleDesktopsURL = "http://simpledesktops.com"
)

// TODO move saving to database logic to upper level.

// SimpleDesktopsProvider is provider of images taken
// from http://simpledesktops.com
type SimpleDesktopsProvider struct {
	config        *config.Config
	storage       *storage.Storage
	maxFetchPages int
}

func (f *SimpleDesktopsProvider) Init(config *config.Config, storage *storage.Storage) {
	log.Println("Initializing SimpleDesktopsProvider...")
	f.config = config
	f.storage = storage
	f.maxFetchPages = f.config.MaxFetchPages
}

// Provide tries to parse and download images from http://simpledesktops.com.
func (f *SimpleDesktopsProvider) Provide() *storage.Wallpaper {
	log.Printf("Fetching from %s", simpleDesktopsURL)

	var wallpaper *storage.Wallpaper

	for wallpaper == nil || wallpaper.ID == 0 {
		rand.Seed(time.Now().UnixNano())
		pageNum := rand.Intn(f.maxFetchPages-1) + 1
		url := fmt.Sprintf(entryPointURL, pageNum)
		log.Printf("Fetching %s...", url)

		wallpaper = f.tryToPickFrom(url)

		// Here maxFetchPages is being approximated to real amount
		// pages on the website on each iteration.

		// Example: The website has 50 pages and maxFetchPages
		// is equal to 50. Imagine we try to parse page #80 on
		// first iteration. After parsing we get wallpaper equal
		// nil. So that means page #80 does not exist (at least
		// has no wallpapers). Consequently we don't need to
		// look at pages 81, 82, 83, etc. Also we (hope that we)
		// can't get 404 error on page #35 if we have 50 pages
		// on website at all.
		if wallpaper == nil && f.maxFetchPages > pageNum {
			f.maxFetchPages = pageNum - 1
		}
	}

	return wallpaper
}

func (f *SimpleDesktopsProvider) tryToPickFrom(url string) *storage.Wallpaper {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[Provide %s] %v", url, err)
		return &storage.Wallpaper{}
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("[Read document at %s] %v", url, err)
		return &storage.Wallpaper{}
	}

	var selectedWallpaper *storage.Wallpaper

	desktops := doc.
		Find(".desktops").
		Find(".edge")

	if desktops.Length() > 0 {
		rand.Seed(time.Now().UnixNano())
		desktopIndex := rand.Intn(desktops.Length())

		pageUrl := ""
		pageUrlExists := false
		title := ""
		author := ""
		authorUrl := ""

		doc.
			Find(".desktops").
			Find(".edge").
			Each(func(i int, selection *goquery.Selection) {
				if i == desktopIndex {
					log.Printf("Parsing HTML element #%d...", i+1)
					wallpaperLink := selection.
						Find(".desktop").
						Find("a")

					pageUrl, pageUrlExists = wallpaperLink.Attr("href")
					title = selection.
						Find(".desktop").
						Find("h2").
						Text()

					authorLink := selection.
						Find(".desktop").
						Find(".creator").
						Find("a")

					author = strings.TrimSpace(authorLink.Text())
					if len(author) == 0 {
						author = "Unknown"
					}
					authorUrl, _ = authorLink.Attr("href")
				}
			})

		if !pageUrlExists {
			log.Printf("Failed to find link wallpaper page on page %s", pageUrl)
			return &storage.Wallpaper{}
		}

		pageUrl = fmt.Sprintf("%s%s", simpleDesktopsURL, pageUrl)

		filename, img, err := pullWallpaperFromPage(pageUrl)
		if err != nil {
			log.Printf("[Provide wallpaper from %s] %v", pageUrl, err)
			return &storage.Wallpaper{}
		}

		wallpaper := &storage.Wallpaper{
			OriginURL:      pageUrl,
			Filename:       filename,
			FetchTimestamp: uint(time.Now().Unix()),
			Title:          title,
			Author:         author,
			AuthorURL:      authorUrl,
		}

		id, err := f.storage.AddWallpaper(wallpaper)
		if err != nil {
			log.Printf("[Save wallpaper to database] %v", err)
			return &storage.Wallpaper{}
		}

		f.saveImage(filename, img)
		wallpaper.ID = id
		selectedWallpaper = wallpaper
	}

	return selectedWallpaper
}

func pullWallpaperFromPage(url string) (string, []byte, error) {
	log.Printf("Fetching image from wallpaper page: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", []byte{}, nil
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", []byte{}, err
	}

	desktopDetail := doc.Find(".desktop-detail").Find(".desktop")
	imgURL, ok := desktopDetail.Find("a").Attr("href")

	if !ok {
		return "", []byte{}, errors.New("failed to extract image url")
	}
	imgURL = fmt.Sprintf("%s%s", simpleDesktopsURL, imgURL)
	log.Printf("Image URL: %s", imgURL)

	filename, img, err := downloadImageToBuffer(imgURL)
	if err != nil {
		return "", []byte{}, err
	}

	return filename, img, nil
}

func downloadImageToBuffer(url string) (string, []byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", []byte{}, err
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()

	fileUUID := uuid.New()
	filename := fmt.Sprintf(
		"%s-%s",
		fileUUID.String(),
		path.Base(finalURL),
	)

	img, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", []byte{}, err
	}

	return filename, img, nil
}

func (f *SimpleDesktopsProvider) saveImage(filename string, image []byte) {
	if _, err := os.Stat(f.config.LocalStoragePath); os.IsNotExist(err) {
		if err := os.MkdirAll(f.config.LocalStoragePath, os.ModePerm); err != nil {
			log.Fatalf(
				"Failed to create images directory %s: %v",
				f.config.LocalStoragePath,
				err,
			)
		}
	}

	filepath := path.Join(f.config.LocalStoragePath, filename)
	if err := ioutil.WriteFile(filepath, image, os.ModePerm); err != nil {
		log.Fatalf("Failed to write image to %s: %v", filepath, err)
	}

	log.Printf("Saved image to %s", filepath)

}

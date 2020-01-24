package fetch

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
	"time"
)

const (
	entryPointURL = "http://simpledesktops.com/browse/%d"
)

type SimpleDesktopsFetcher struct {
	config  *config.Config
	storage *storage.Storage
}

func (f *SimpleDesktopsFetcher) Init(config *config.Config, storage *storage.Storage) {
	log.Println("Initializing SimpleDesktopsFetcher...")
	f.config = config
	f.storage = storage
}

// Fetch tries to parse and download images from http://simpledesktops.com.
// Parameter limit means max count of pages to visit.
func (f *SimpleDesktopsFetcher) Fetch() *storage.Wallpaper {
	log.Println("Fetching from http://simpledesktops.com...")

	var wallpaper *storage.Wallpaper

	for wallpaper == nil || wallpaper.ID == 0 {
		rand.Seed(time.Now().UnixNano())
		pageNum := rand.Intn(f.config.MaxFetchPages)
		url := fmt.Sprintf(entryPointURL, pageNum)
		log.Printf("Fetching %s...", url)

		wallpaper = f.fetchFromOrigin(url)

		time.Sleep(time.Duration(f.config.SleepTime) * time.Second)
		//})
	}
	//l.Wait()

	return wallpaper
}

func (f *SimpleDesktopsFetcher) fetchFromOrigin(url string) *storage.Wallpaper {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[Fetch %s] %v", url, err)
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

		doc.
			Find(".desktops").
			Find(".edge").
			Each(func(i int, selection *goquery.Selection) {
				if i == desktopIndex {
					log.Printf("Parsing HTML element #%d...", (i + 1))
					pageURL, exists := selection.
						Find(".desktop").
						Find("a").
						Attr("href")

					if !exists {
						log.Printf("Failed to find link wallpaper page on page %s", pageURL)
						return
					}

					pageURL = fmt.Sprintf("http://simpledesktops.com/%s", pageURL)

					filename, img, err := fetchWallpaperFromPage(pageURL)
					if err != nil {
						log.Printf("[Fetch wallpaper from %s] %v", pageURL, err)
						return
					}

					wallpaper := &storage.Wallpaper{
						OriginURL:      pageURL,
						Filename:       filename,
						FetchTimestamp: uint(time.Now().Unix()),
					}

					id, err := f.storage.AddWallpaper(wallpaper)
					if err != nil {
						log.Printf("[Save wallpaper to database] %v", err)
						return
					}

					f.saveImage(filename, img)
					wallpaper.ID = id
					selectedWallpaper = wallpaper
				}
			})
	}

	return selectedWallpaper
}

func fetchWallpaperFromPage(url string) (string, []byte, error) {
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
	imgURL = fmt.Sprintf("http://simpledesktops.com%s", imgURL)
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

func (f *SimpleDesktopsFetcher) saveImage(filename string, image []byte) {
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
	log.Printf("Saving image to %s...", filepath)

	if err := ioutil.WriteFile(filepath, image, os.ModePerm); err != nil {
		log.Fatalf("Failed to write image to %s: %v", filepath, err)
	}

	log.Printf("Saved image to %s", filepath)

}

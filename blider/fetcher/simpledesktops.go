package fetcher

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ildarkarymoff/blider/blider"
	"github.com/ildarkarymoff/blider/blider/storage"
	"github.com/zenthangplus/goccm"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

const (
	entryPointURL = "http://simpledesktops.com/browse/%d"
)

type SimpleDesktopsFetcher struct {
	config     *blider.Config
	storage    *storage.Storage
	wallpapers []*storage.Wallpaper
}

func (f SimpleDesktopsFetcher) Init(config *blider.Config) {
	f.config = config
}

// Fetch tries to parse and download images from http://simpledesktops.com.
// Parameter limit means max count of pages to visit.
func (f SimpleDesktopsFetcher) Fetch(limit int) []*storage.Wallpaper {
	f.wallpapers = []*storage.Wallpaper{}

	manager := goccm.New(f.config.MaxFetchGoroutines)
	pageCounter := 1

	for pageCounter < limit {
		manager.Wait()
		go func() {
			url := fmt.Sprintf(entryPointURL, pageCounter)
			f.fetchFromOrigin(url)

			manager.Done()
		}()

		pageCounter++
	}

	manager.WaitAllDone()
	return f.wallpapers
}

func (f *SimpleDesktopsFetcher) fetchFromOrigin(url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[Fetch %s] %v", url, err)
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("[Read document at %s] %v", url, err)
		return
	}

	doc.
		Find(".desktops .column .span-24 .archive").
		Find(".edge").
		Each(func(i int, selection *goquery.Selection) {
			pageURL, exists := selection.
				Find(".desktop").
				Find("a").
				Attr("href")

			if !exists {
				return
			}

			filename, img, err := fetchWallpaperFromPage(pageURL)
			if err != nil {
				log.Printf("[Fetch wallpaper from %s] %v", pageURL, err)
			}

			f.saveImage(filename, img)
			f.wallpapers = append(f.wallpapers, &storage.Wallpaper{
				OriginURL:      pageURL,
				Filename:       filename,
				FetchTimestamp: uint(time.Now().Unix()),
			})
		})

}

func fetchWallpaperFromPage(url string) (string, []byte, error) {
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

	finalURL := resp.Request.URL
	filename := path.Base(finalURL.Path)

	img, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", []byte{}, err
	}

	return filename, img, nil
}

func (f *SimpleDesktopsFetcher) saveImage(filename string, image []byte) {
	filepath := path.Join(f.config.LocalStoragePath, filename)
	if err := ioutil.WriteFile(filepath, image, os.ModePerm); err != nil {
		log.Fatalf("failed to write image to %s", filename)
	}
}

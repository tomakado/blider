package provider

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/ildarkarymoff/blider/config"
	"github.com/ildarkarymoff/blider/repository"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"path"
	"strings"
	"time"
)

const (
	entryPointURL     = "http://simpledesktops.com/browse/%d"
	simpleDesktopsURL = "http://simpledesktops.com"
)

// SimpleDesktopsProvider is provider of images taken
// from http://simpledesktops.com
type SimpleDesktopsProvider struct {
	config        *config.Config
	repository    *repository.Repository
	maxFetchPages int
}

func (p *SimpleDesktopsProvider) Init(cfg *config.Config, repository *repository.Repository) {
	log.Println("Initializing SimpleDesktopsProvider...")
	p.config = cfg
	p.repository = repository
	p.maxFetchPages = 10

	providerConfig, cfgOk := p.config.Providers[config.ProviderSimpleDesktops]
	if cfgOk {
		maxFetchPages, mfpOk := (*providerConfig)["max_fetch_pages"].(int)
		if mfpOk {
			p.maxFetchPages = func(a, b int) int {
				if a > b {
					return a
				}

				return b
			}(maxFetchPages, 0)
		}
	}
}

// Provide tries to parse and download images from http://simpledesktops.com.
func (p *SimpleDesktopsProvider) Provide() *repository.Wallpaper {
	log.Printf("Fetching from %s...", simpleDesktopsURL)

	var wallpaper *repository.Wallpaper

	for wallpaper == nil {
		rand.Seed(time.Now().UnixNano())
		pageNum := rand.Intn(p.maxFetchPages-1) + 1
		url := fmt.Sprintf(entryPointURL, pageNum)
		log.Printf("Fetching %s...", url)

		wallpaper = p.tryToPickFrom(url)

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
		if wallpaper == nil && p.maxFetchPages > pageNum {
			p.maxFetchPages = pageNum - 1
		}
	}

	return wallpaper
}

func (p *SimpleDesktopsProvider) tryToPickFrom(url string) *repository.Wallpaper {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[Provide %s] %v", url, err)
		return &repository.Wallpaper{}
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("[Read document at %s] %v", url, err)
		return &repository.Wallpaper{}
	}

	var selectedWallpaper *repository.Wallpaper

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
				if i != desktopIndex {
					return
				}

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
			})

		if !pageUrlExists {
			log.Printf("Failed to find link wallpaper page on page %s", url)
			return &repository.Wallpaper{}
		}

		pageUrl = fmt.Sprintf("%s%s", simpleDesktopsURL, pageUrl)

		filename, img, err := pullWallpaperFromPage(pageUrl)
		if err != nil {
			log.Printf("[Provide wallpaper from %s] %v", pageUrl, err)
			return &repository.Wallpaper{}
		}

		wallpaper := &repository.Wallpaper{
			OriginURL:      pageUrl,
			Filename:       filename,
			FetchTimestamp: uint(time.Now().Unix()),
			Title:          title,
			Author:         author,
			AuthorURL:      authorUrl,
			ImgBuffer:      img,
		}

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

	imgSize := float32(len(img))

	log.Printf(
		"Downloaded '%s' / %.2f KB",
		filename,
		imgSize/1024,
	)
	return filename, img, nil
}

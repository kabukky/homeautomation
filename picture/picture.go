package picture

import (
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/kabukky/homeautomation/utils"
	"github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
)

var (
	randInstance                = rand.New(rand.NewSource(time.Now().UnixNano()))
	cachedPicturesOfTheDay      = map[string]pictureOfTheDay{} // Key: deviceID
	cachedPicturesOfTheDayMutex sync.RWMutex
)

type pictureOfTheDay struct {
	generatedAt time.Time
	filename    string
}

func GetPictureOfTheDay(deviceID string) (image.Image, error) {
	cachedPicturesOfTheDayMutex.RLock()
	entry, ok := cachedPicturesOfTheDay[deviceID]
	cachedPicturesOfTheDayMutex.RUnlock()
	if !ok || time.Now().After(entry.generatedAt.Add(1*time.Hour)) {
		// Entry not generated yet
		// Or generated > 1 hours ago
		// Generate a new entry
		filename, err := getRandomFilename()
		if err != nil {
			return nil, err
		}
		entry = pictureOfTheDay{
			generatedAt: time.Now(),
			filename:    filename,
		}
		cachedPicturesOfTheDayMutex.Lock()
		cachedPicturesOfTheDay[deviceID] = entry
		cachedPicturesOfTheDayMutex.Unlock()
	}
	// Read file
	f, err := os.Open(entry.filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := imaging.Decode(f, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	// Adjust saturation for eink
	img = imaging.AdjustSaturation(img, 50)
	return crop(img, 0, 0, true)
}

func getRandomFilename() (string, error) {
	root := os.DirFS(utils.PicturesOfTheDayDirectory)
	// Get all images
	var allMatches []string
	matches, err := fs.Glob(root, "*.jpg")
	if err != nil {
		return "", err
	}
	allMatches = append(allMatches, matches...)
	matches, err = fs.Glob(root, "*.jpeg")
	if err != nil {
		return "", err
	}
	allMatches = append(allMatches, matches...)
	matches, err = fs.Glob(root, "*.png")
	if err != nil {
		return "", err
	}
	allMatches = append(allMatches, matches...)
	if len(allMatches) == 0 {
		return "", errors.New("no matching files in directory")
	}
	index := randInstance.Intn(len(allMatches))
	return filepath.Join(utils.PicturesOfTheDayDirectory, allMatches[index]), nil
}

func crop(img image.Image, w, h int, resize bool) (image.Image, error) {
	width, height := getCropDimensions(img, w, h)
	resizer := nfnt.NewDefaultResizer()
	analyzer := smartcrop.NewAnalyzer(resizer)
	topCrop, err := analyzer.FindBestCrop(img, width, height)
	if err != nil {
		return nil, err
	}

	type SubImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	img = img.(SubImager).SubImage(topCrop)
	if resize && (img.Bounds().Dx() != width || img.Bounds().Dy() != height) {
		img = resizer.Resize(img, uint(width), uint(height))
	}
	return img, nil
}

func getCropDimensions(img image.Image, width, height int) (int, int) {
	// if we don't have width or height set use the smaller image dimension as both width and height
	if width == 0 && height == 0 {
		bounds := img.Bounds()
		x := bounds.Dx()
		y := bounds.Dy()
		if x < y {
			width = x
			height = x
		} else {
			width = y
			height = y
		}
	}
	return width, height
}

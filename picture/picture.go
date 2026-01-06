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
	return img, nil
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
